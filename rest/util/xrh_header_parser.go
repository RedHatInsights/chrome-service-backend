package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/sirupsen/logrus"
)

func ParseXRHIdentityHeader(identityHeader string) (*identity.XRHID, error) {
	var XRHIdentity identity.XRHID
	decodedIdentity, err := base64.StdEncoding.DecodeString(identityHeader)
	if err != nil {
		return nil, fmt.Errorf("error decoding Identity: %v", err)
	}

	err = json.Unmarshal(decodedIdentity, &XRHIdentity)

	if err != nil {
		err = fmt.Errorf("x-rh-identity header is not a valid json: %w", err)
		logrus.Error(err)
		return nil, err
	}

	// XRHIdentity.Identity.User.UserID
	return &XRHIdentity, nil
}

