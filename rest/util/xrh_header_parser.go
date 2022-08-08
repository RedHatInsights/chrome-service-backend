package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/sirupsen/logrus"
)

func ParseXRHIdentityheader(identityHeader string) (*identity.XRHID, error) {
	var XRHIdentity identity.XRHID
	decodedIdentity, err := base64.StdEncoding.DecodeString(identityHeader)
	if err != nil {
		return nil, fmt.Errorf("error decoding Identity: %v", err)
	}

	err = json.Unmarshal(decodedIdentity, &XRHIdentity)

	if err != nil {
		logrus.Errorf("x-rh-identity header is not a valid json: %s. Identity: %s", err.Error(), identityHeader)
		return nil, fmt.Errorf("x-rh-identity header is not a valid json: %s", err.Error())
	}

	// XRHIdentity.Identity.User.UserID
	return &XRHIdentity, nil
}
