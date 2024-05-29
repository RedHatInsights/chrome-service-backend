package util

import (
	"errors"
)

const (
	XRHIDENTITY      = "x-rh-identity"
	LAST_VISITED_MAX = 10
	IDENTITY_CTX_KEY = "identity"
	USER_CTX_KEY     = "user"
	GET_ALL_PARAM    = "getAll"   // Used for searching ALL favorited pages
	DEFAULT_PARAM    = "archived" // Used as default value for active favorited pages
	FAVORITE_PARAM   = "favorite"
)

var (
	ErrNotAuthorized = errors.New("not authorized")
	ErrBadRequest    = errors.New("bad request")
)
