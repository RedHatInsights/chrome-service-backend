package util_test

import (
	"testing"
	"time"

	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	util.InitUserIdentitiesCache()
	t.Run("Test Cache Set", func(t *testing.T) {
		testIdentity := models.UserIdentity{
			AccountId: "test",
		}
		util.UsersCache.Set("test", testIdentity)

		identity, ok := util.UsersCache.Get("test")
		assert.True(t, ok)
		assert.Equal(t, "test", identity.AccountId)
		assert.Equal(t, identity, testIdentity)
	})

	t.Run("Test Cache Delete", func(t *testing.T) {
		util.UsersCache.Delete("test")
		identity, ok := util.UsersCache.Get("test")
		assert.False(t, ok)
		assert.Equal(t, models.UserIdentity{}, identity)
	})

	t.Run("Test Cache Expire", func(t *testing.T) {
		testIdentity := models.UserIdentity{
			AccountId: "expired",
		}
		util.UsersCache.Set("expired", testIdentity)

		identity, ok := util.UsersCache.Get("expired")
		assert.True(t, ok)
		assert.Equal(t, "expired", identity.AccountId)
		assert.Equal(t, identity, testIdentity)

		// Expire the cache by force
		util.UsersCache.Identities["expired"] = util.CacheEntry{
			ExpireAt: util.UsersCache.Identities["expired"].ExpireAt.Add(-time.Minute * 2),
			Identity: testIdentity,
		}

		identity, ok = util.UsersCache.Get("expired")
		assert.False(t, ok)
		assert.Equal(t, models.UserIdentity{}, identity)
	})
}
