package util

import (
	"sync"
	"time"

	"github.com/RedHatInsights/chrome-service-backend/rest/models"
)

type CacheEntry struct {
	ExpireAt time.Time
	Identity models.UserIdentity
}

type UserIdentityCache struct {
	Identities map[string]CacheEntry
	sync.Mutex
}

func (c *UserIdentityCache) Get(accountId string) (models.UserIdentity, bool) {
	// Lock the cache to prevent out of sync operations
	c.Lock()
	defer c.Unlock()

	entry, ok := c.Identities[accountId]
	if !ok {
		return models.UserIdentity{}, false
	}
	if entry.ExpireAt.Before(time.Now()) {
		delete(c.Identities, accountId)
		return models.UserIdentity{}, false
	}
	return entry.Identity, true
}

func (c *UserIdentityCache) Set(accountId string, identity models.UserIdentity) {
	c.Lock()
	defer c.Unlock()

	c.Identities[accountId] = CacheEntry{
		/**
		* Fairly short cache time is OK
		* Mainly we are looking to speed up burst traffic for a single user
		* Typically on session start where we get about 3-4 hits per user per session
		 */
		ExpireAt: time.Now().Add(time.Second * 30),
		Identity: identity,
	}
}

func (c *UserIdentityCache) Delete(accountId string) {
	c.Lock()
	defer c.Unlock()

	delete(c.Identities, accountId)
}

var UsersCache *UserIdentityCache

func InitUserIdentitiesCache() {
	UsersCache = &UserIdentityCache{Identities: make(map[string]CacheEntry)}
}
