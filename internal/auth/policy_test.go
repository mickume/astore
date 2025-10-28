package auth_test

import (
	"testing"

	"github.com/candlekeep/zot-artifact-store/internal/auth"
	"github.com/candlekeep/zot-artifact-store/internal/models"
	"github.com/candlekeep/zot-artifact-store/test"
)

func TestPolicyEngine(t *testing.T) {
	t.Run("Admin has full access", func(t *testing.T) {
		// Given: A policy engine and an admin user
		engine := auth.NewPolicyEngine(false)
		user := &models.User{
			ID:       "admin-1",
			Username: "admin",
			Roles:    []string{"admin"},
		}

		// When: Checking authorization for any action
		allowed, err := engine.Authorize(user, "any-bucket", models.ActionWrite)

		// Then: Access is granted
		test.AssertNoError(t, err, "admin authorization")
		test.AssertTrue(t, allowed, "admin has access")
	})

	t.Run("Policy allows specific user access", func(t *testing.T) {
		// Given: A policy engine with a policy
		engine := auth.NewPolicyEngine(false)
		policy := &models.Policy{
			ID:          "policy-1",
			Resource:    "mybucket",
			Actions:     []string{"read", "write"},
			Effect:      models.PolicyEffectAllow,
			Principals:  []string{"user-1"},
		}
		engine.AddPolicy(policy)

		user := &models.User{
			ID:       "user-1",
			Username: "testuser",
		}

		// When: User accesses allowed resource
		allowed, err := engine.Authorize(user, "mybucket", models.ActionRead)

		// Then: Access is granted
		test.AssertNoError(t, err, "policy authorization")
		test.AssertTrue(t, allowed, "user has access")
	})

	t.Run("Policy denies access to different resource", func(t *testing.T) {
		// Given: A policy engine with a policy for specific resource
		engine := auth.NewPolicyEngine(false)
		policy := &models.Policy{
			ID:          "policy-1",
			Resource:    "mybucket",
			Actions:     []string{"read"},
			Effect:      models.PolicyEffectAllow,
			Principals:  []string{"user-1"},
		}
		engine.AddPolicy(policy)

		user := &models.User{
			ID:       "user-1",
			Username: "testuser",
		}

		// When: User accesses different resource
		allowed, _ := engine.Authorize(user, "otherbucket", models.ActionRead)

		// Then: Access is denied
		test.AssertFalse(t, allowed, "no access to other resources")
	})

	t.Run("Wildcard resource allows all", func(t *testing.T) {
		// Given: A policy with wildcard resource
		engine := auth.NewPolicyEngine(false)
		policy := &models.Policy{
			ID:          "policy-1",
			Resource:    "*",
			Actions:     []string{"read"},
			Effect:      models.PolicyEffectAllow,
			Principals:  []string{"user-1"},
		}
		engine.AddPolicy(policy)

		user := &models.User{
			ID:       "user-1",
			Username: "testuser",
		}

		// When: User accesses any resource
		allowed, _ := engine.Authorize(user, "anybucket", models.ActionRead)

		// Then: Access is granted
		test.AssertTrue(t, allowed, "wildcard allows all resources")
	})

	t.Run("Deny policy takes precedence", func(t *testing.T) {
		// Given: Both allow and deny policies
		engine := auth.NewPolicyEngine(false)

		allowPolicy := &models.Policy{
			ID:          "allow-1",
			Resource:    "mybucket",
			Actions:     []string{"read"},
			Effect:      models.PolicyEffectAllow,
			Principals:  []string{"user-1"},
		}
		engine.AddPolicy(allowPolicy)

		denyPolicy := &models.Policy{
			ID:          "deny-1",
			Resource:    "mybucket",
			Actions:     []string{"read"},
			Effect:      models.PolicyEffectDeny,
			Principals:  []string{"user-1"},
		}
		engine.AddPolicy(denyPolicy)

		user := &models.User{
			ID:       "user-1",
			Username: "testuser",
		}

		// When: Checking authorization
		allowed, _ := engine.Authorize(user, "mybucket", models.ActionRead)

		// Then: Access is denied
		test.AssertFalse(t, allowed, "deny takes precedence")
	})

	t.Run("Anonymous GET allowed when configured", func(t *testing.T) {
		// Given: Policy engine with anonymous GET enabled
		engine := auth.NewPolicyEngine(true)

		// When: Anonymous user tries to read
		allowed, err := engine.Authorize(nil, "anybucket", models.ActionRead)

		// Then: Access is granted
		test.AssertNoError(t, err, "anonymous read")
		test.AssertTrue(t, allowed, "anonymous read allowed")
	})

	t.Run("Anonymous write denied", func(t *testing.T) {
		// Given: Policy engine with anonymous GET enabled
		engine := auth.NewPolicyEngine(true)

		// When: Anonymous user tries to write
		allowed, _ := engine.Authorize(nil, "anybucket", models.ActionWrite)

		// Then: Access is denied
		test.AssertFalse(t, allowed, "anonymous write denied")
	})
}
