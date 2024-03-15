package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"reflect"
	"time"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/sirupsen/logrus"

	"gorm.io/datatypes"
)

type IntercomApp string
type IntercomPayload struct {
	Prod string `json:"prod,omitempty"`
	Dev  string `json:"dev,omitempty"`
}

const (
	OpenShift           IntercomApp = "openshift"
	HacCore             IntercomApp = "hacCore"
	Acs                 IntercomApp = "acs"
	Ansible             IntercomApp = "ansible"
	AnsibleDashboard    IntercomApp = "ansibleDashboard"
	AutomationHub       IntercomApp = "automationHub"
	AutomationAnalytics IntercomApp = "automationAnalytics"
	DBAAS               IntercomApp = "dbaas"
)

func debugFavoritesIdentity(userId string) {
	c := config.Get()
	for _, i := range c.DebugConfig.DebugFavoriteIds {
		if i == userId {
			logrus.Warningln("DEBUG_FAVORITES_ACCOUNT_ID", userId)
		}
	}
}

func (ib IntercomApp) IsValidApp() error {
	switch ib {
	case OpenShift, HacCore, Ansible, Acs, AnsibleDashboard, AutomationHub, AutomationAnalytics, DBAAS:
		return nil
	}

	return fmt.Errorf("invalid bundle string. Expected one of %s, %s, got %s", OpenShift, HacCore, ib)
}

func parseUserBundles(user models.UserIdentity) (map[string]bool, error) {
	bundles := make(map[string]bool)
	// make sure not to potentially marshal nil to map
	if user.VisitedBundles == nil {
		return bundles, nil
	}
	err := json.Unmarshal(user.VisitedBundles, &bundles)
	return bundles, err
}

// Get user data complete with it's related tables.
func GetUserIdentityData(user models.UserIdentity) (models.UserIdentity, error) {
	var favoritePages []models.FavoritePage
	err := database.DB.Model(&user).Association("FavoritePages").Find(&favoritePages)
	debugFavoritesIdentity(user.AccountId)

	user.FavoritePages = favoritePages
	return user, err
}

// Set visited bundle
func AddVisitedBundle(user models.UserIdentity, bundle string) (models.UserIdentity, error) {
	bundles, err := parseUserBundles(user)
	if err != nil {
		return models.UserIdentity{}, err
	}
	// if the bundles object does not exist create it
	if bundles == nil {
		bundles = make(map[string]bool)
		err := json.Unmarshal([]byte(`{}`), &bundles)
		if err != nil {
			return user, err
		}
	}
	bundles[bundle] = true
	b, err := json.Marshal(bundles)
	if err != nil {
		return models.UserIdentity{}, err
	}
	// update the bundle reference for the function scope
	user.VisitedBundles = b
	err = database.DB.Model(&user).Update("visited_bundles", bundles).Error
	return user, err
}

func GetVisitedBundles(user models.UserIdentity) (map[string]bool, error) {
	return parseUserBundles(user)
}

// Create the user object and add the row if not already in DB
func CreateIdentity(userId string) (models.UserIdentity, error) {
	identity := models.UserIdentity{
		AccountId:        userId,
		FirstLogin:       true,
		DayOne:           true,
		LastLogin:        time.Now(),
		LastVisitedPages: datatypes.NewJSONType([]models.VisitedPage{}),
		FavoritePages:    []models.FavoritePage{},
		SelfReport:       models.SelfReport{},
		VisitedBundles:   nil,
	}
	err := json.Unmarshal([]byte(`{}`), &identity.VisitedBundles)
	if err != nil {
		return models.UserIdentity{}, err
	}

	/**
	* Because we pass the object from the middleware to the rest of the application,
	* we don't have to worry about invalidation the cache, as the object is passed by reference
	* saves a lot DB queries.
	 */
	cachedIdentity, ok := util.UsersCache.Get(userId)
	if ok {
		return cachedIdentity, nil
	}

	res := database.DB.Where("account_id = ?", userId).FirstOrCreate(&identity)
	err = res.Error

	// set the cache after successful DB operation
	if err == nil {
		util.UsersCache.Set(userId, identity)
	}

	return identity, err
}

func encodeKey(namespace string, userId string) (string, error) {
	var intercomHash hash.Hash
	var err error
	c := config.Get()
	v := reflect.ValueOf(c.IntercomConfig)
	key := reflect.Indirect(v).FieldByName(string(namespace))
	if key.IsValid() {
		intercomHash = hmac.New(sha256.New, []byte(key.String()))
		_, err = intercomHash.Write([]byte(userId))
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(intercomHash.Sum(nil)), nil
	}

	// is not a valid key, do not encode
	return "", nil
}

func GetUserIntercomHash(userId string, namespace IntercomApp) (IntercomPayload, error) {
	err := namespace.IsValidApp()
	response := IntercomPayload{}
	if err != nil {
		logrus.Infof("Unable to verify intercom namespace %s", namespace)
		return response, nil
	}
	devNamespace := fmt.Sprintf("%s_dev", namespace)

	prodKey, err := encodeKey(string(namespace), userId)
	if err != nil {
		return response, err
	}
	response.Prod = prodKey

	devKey, err := encodeKey(devNamespace, userId)

	if err != nil {
		return response, err
	}
	response.Dev = devKey
	return response, nil
}
