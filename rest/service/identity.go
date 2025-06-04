package service

import (
	"context"
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
	redis_client "github.com/RedHatInsights/chrome-service-backend/rest/redis"
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
	ActivationKeys      IntercomApp = "activationKeys"
	Advisor             IntercomApp = "advisor"
	Compliance          IntercomApp = "compliance"
	Connector           IntercomApp = "connector"
	ContentSources      IntercomApp = "contentSources"
	Dashboard           IntercomApp = "dashboard"
	ImageBuilder        IntercomApp = "imageBuilder"
	Inventory           IntercomApp = "inventory"
	Malware             IntercomApp = "malware"
	Patch               IntercomApp = "patch"
	Policies            IntercomApp = "policies"
	Registration        IntercomApp = "registration"
	Remediations        IntercomApp = "remediations"
	Ros                 IntercomApp = "ros"
	Tasks               IntercomApp = "tasks"
	Vulnerability       IntercomApp = "vulnerability"
)

func getIdentityFromCache(userId string) (models.UserIdentity, error) {
	rc := redis_client.GetRedisClient()
	ctx := context.Background()
	var user models.UserIdentity

	val, err := rc.Get(ctx, userId).Result()
	if err != nil {
		return models.UserIdentity{}, fmt.Errorf("error getting user identity from cache: %w", err)
	}
	err = json.Unmarshal([]byte(val), &user)
	return user, err

}

func writeIdentityToCache(user models.UserIdentity) error {
	rc := redis_client.GetRedisClient()
	ctx := context.Background()
	userB, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("error serializing user identity: %w", err)
	}
	err = rc.Set(ctx, user.AccountId, string(userB), time.Minute*10).Err()

	return err
}

func invalidateIdentityCache(userId string) error {
	rc := redis_client.GetRedisClient()
	ctx := context.Background()
	return rc.Del(ctx, userId).Err()
}

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
	case OpenShift, HacCore, Ansible, Acs, AnsibleDashboard, AutomationHub, AutomationAnalytics, DBAAS, ActivationKeys, Advisor, Compliance, Connector, ContentSources, Dashboard, ImageBuilder, Inventory, Malware, Patch, Policies, Registration, Remediations, Ros, Tasks, Vulnerability:
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

	if err != nil {
		logrus.Errorf("Error updating user preview: %v", err)
		return user, err
	}
	err = invalidateIdentityCache(user.AccountId)
	if err != nil {
		logrus.Errorf("Error writing identity to cache after updating preview: %v", err)
	}
	return user, nil
}

func GetVisitedBundles(user models.UserIdentity) (map[string]bool, error) {
	return parseUserBundles(user)
}

// Create the user object and add the row if not already in DB
func CreateIdentity(userId string, skipCache bool) (models.UserIdentity, error) {
	identity := models.UserIdentity{
		AccountId:        userId,
		FirstLogin:       true,
		DayOne:           true,
		LastLogin:        time.Now(),
		LastVisitedPages: datatypes.NewJSONType([]models.VisitedPage{}),
		FavoritePages:    []models.FavoritePage{},
		SelfReport:       models.SelfReport{},
		VisitedBundles:   nil,
		ActiveWorkspace:  "default",
	}
	err := json.Unmarshal([]byte(`{}`), &identity.VisitedBundles)
	if err != nil {
		return models.UserIdentity{}, err
	}

	/**
	* Because we pass the object from the middleware to the rest of the application,
	* saves a lot DB queries.
	 */
	cachedIdentity, err := getIdentityFromCache(userId)
	if err == nil {
		return cachedIdentity, nil
	} else {
		logrus.Errorf("Error getting user identity from cache: %v", err)
	}

	res := database.DB.Where("account_id = ?", userId).FirstOrCreate(&identity)
	err = res.Error

	// set the cache after successful DB operation
	if err == nil {
		err = writeIdentityToCache(identity)
		if err != nil {
			logrus.Errorf("Error writing identity to cache: %v", err)
		}
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

func UpdateUserPreview(identity *models.UserIdentity, preview bool) error {
	identity.UIPreview = preview
	err := database.DB.Model(identity).Update("ui_preview", preview).Error
	if err != nil {
		logrus.Errorf("Error updating user preview: %v", err)
		return err
	}

	err = invalidateIdentityCache(identity.AccountId)
	if err != nil {
		logrus.Errorf("Error writing identity to cache after updating preview: %v", err)
	}
	return nil
}

func MarkPreviewSeen(identity *models.UserIdentity) error {
	err := database.DB.Model(identity).Updates(models.UserIdentity{UIPreviewSeen: true}).Error
	if err != nil {
		logrus.Errorf("Error updating user preview: %v", err)
		return err
	}
	err = invalidateIdentityCache(identity.AccountId)
	if err != nil {
		logrus.Errorf("Error writing identity to cache after updating preview: %v", err)
	}
	return nil
}

func UpdateActiveWorkspace(identity *models.UserIdentity, workspace string) error {
	identity.ActiveWorkspace = workspace
	err := database.DB.Model(identity).Update("active_workspace", workspace).Error

	if err != nil {
		logrus.Errorf("Error updating user preview: %v", err)
		return err
	}
	err = invalidateIdentityCache(identity.AccountId)
	if err != nil {
		logrus.Errorf("Error writing identity to cache after updating preview: %v", err)
	}
	return nil
}
