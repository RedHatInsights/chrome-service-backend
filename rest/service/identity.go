package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/sirupsen/logrus"
)

type IntercomBundle string

const (
	Fallback IntercomBundle = "fallback"
)

func (ib IntercomBundle) IsValidBundle() error {
	switch ib {
	case Fallback:
		return nil
	}

	return fmt.Errorf("invalid bundle string. Expected one of %s, got %s", Fallback, ib)
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
	var lastVisitedPages []models.LastVisitedPage
	var favoritePages []models.FavoritePage

	err := database.DB.Model(&user).Association("LastVisitedPages").Find(&lastVisitedPages)
	if err != nil {
		return user, err
	}
	err = database.DB.Model(&user).Association("FavoritePages").Find(&favoritePages)

	user.LastVisitedPages = lastVisitedPages
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
		LastVisitedPages: []models.LastVisitedPage{},
		FavoritePages:    []models.FavoritePage{},
		SelfReport:       models.SelfReport{},
		VisitedBundles:   nil,
	}
	err := json.Unmarshal([]byte(`{}`), &identity.VisitedBundles)
	if err != nil {
		return models.UserIdentity{}, err
	}

	res := database.DB.Where("account_id = ?", userId).FirstOrCreate(&identity)
	err = res.Error

	return identity, err
}

func GetUserIntercomHash(userId string, namespace IntercomBundle) (string, error) {
	err := namespace.IsValidBundle()
	bundle := namespace
	if err != nil {
		logrus.Infof("Unable to verify intercom namespace %s. Using fallback key.\n", string(namespace))
		bundle = Fallback
	}
	// get this form env
	c := config.Get()
	// access struct value via string variable
	v := reflect.ValueOf(c.IntercomConfig)
	key := reflect.Indirect(v).FieldByName(string(bundle))

	fmt.Print(key.String(), namespace)
	intercomHash := hmac.New(sha256.New, []byte(key.String()))
	_, err = intercomHash.Write([]byte(userId))
	if err != nil {
		return "", err
	}

	hash := hex.EncodeToString(intercomHash.Sum(nil))
	return hash, nil
}
