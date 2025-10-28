package auth

import (
	"fmt"
	"strings"

	"github.com/candlekeep/zot-artifact-store/internal/models"
)

// PolicyEngine evaluates access control policies
type PolicyEngine struct {
	policies          []*models.Policy
	allowAnonymousGet bool
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine(allowAnonymousGet bool) *PolicyEngine {
	return &PolicyEngine{
		policies:          make([]*models.Policy, 0),
		allowAnonymousGet: allowAnonymousGet,
	}
}

// AddPolicy adds a policy to the engine
func (e *PolicyEngine) AddPolicy(policy *models.Policy) {
	e.policies = append(e.policies, policy)
}

// RemovePolicy removes a policy by ID
func (e *PolicyEngine) RemovePolicy(policyID string) {
	for i, p := range e.policies {
		if p.ID == policyID {
			e.policies = append(e.policies[:i], e.policies[i+1:]...)
			return
		}
	}
}

// ListPolicies returns all policies
func (e *PolicyEngine) ListPolicies() []*models.Policy {
	return e.policies
}

// Authorize checks if a user is authorized to perform an action on a resource
func (e *PolicyEngine) Authorize(user *models.User, resource string, action models.Action) (bool, error) {
	// Handle anonymous access for GET operations
	if user == nil {
		if e.allowAnonymousGet && action == models.ActionRead {
			return true, nil
		}
		return false, fmt.Errorf("authentication required")
	}

	// Admin role has full access
	if e.hasRole(user, "admin") {
		return true, nil
	}

	// Check policies
	allowCount := 0
	denyCount := 0

	for _, policy := range e.policies {
		// Check if policy applies to this resource
		if !e.matchesResource(policy.Resource, resource) {
			continue
		}

		// Check if policy applies to this action
		if !e.containsAction(policy.Actions, string(action)) {
			continue
		}

		// Check if policy applies to this principal
		if !e.matchesPrincipal(policy.Principals, user) {
			continue
		}

		// Apply policy effect
		switch policy.Effect {
		case models.PolicyEffectAllow:
			allowCount++
		case models.PolicyEffectDeny:
			denyCount++
		}
	}

	// Deny takes precedence
	if denyCount > 0 {
		return false, fmt.Errorf("access denied by policy")
	}

	// If any policy allows, grant access
	if allowCount > 0 {
		return true, nil
	}

	// Default deny
	return false, fmt.Errorf("no policy allows this action")
}

// GetPermissions returns all permissions for a user
func (e *PolicyEngine) GetPermissions(user *models.User) []models.Permission {
	if user == nil {
		if e.allowAnonymousGet {
			return []models.Permission{
				{Resource: "*", Actions: []models.Action{models.ActionRead}},
			}
		}
		return []models.Permission{}
	}

	// Admin has all permissions
	if e.hasRole(user, "admin") {
		return []models.Permission{
			{Resource: "*", Actions: []models.Action{
				models.ActionRead,
				models.ActionWrite,
				models.ActionDelete,
				models.ActionList,
				models.ActionAdmin,
			}},
		}
	}

	// Collect permissions from policies
	permMap := make(map[string]map[models.Action]bool)

	for _, policy := range e.policies {
		if policy.Effect != models.PolicyEffectAllow {
			continue
		}

		if !e.matchesPrincipal(policy.Principals, user) {
			continue
		}

		if _, ok := permMap[policy.Resource]; !ok {
			permMap[policy.Resource] = make(map[models.Action]bool)
		}

		for _, actionStr := range policy.Actions {
			action := models.Action(actionStr)
			permMap[policy.Resource][action] = true
		}
	}

	// Convert map to slice
	permissions := make([]models.Permission, 0, len(permMap))
	for resource, actions := range permMap {
		actionList := make([]models.Action, 0, len(actions))
		for action := range actions {
			actionList = append(actionList, action)
		}
		permissions = append(permissions, models.Permission{
			Resource: resource,
			Actions:  actionList,
		})
	}

	return permissions
}

// matchesResource checks if a policy resource pattern matches the actual resource
func (e *PolicyEngine) matchesResource(pattern, resource string) bool {
	// Exact match
	if pattern == resource {
		return true
	}

	// Wildcard match
	if pattern == "*" {
		return true
	}

	// Prefix match for bucket patterns (e.g., "mybucket/*")
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		if strings.HasPrefix(resource, prefix+"/") || resource == prefix {
			return true
		}
	}

	return false
}

// containsAction checks if an action list contains a specific action
func (e *PolicyEngine) containsAction(actions []string, action string) bool {
	for _, a := range actions {
		if a == action || a == "*" {
			return true
		}
	}
	return false
}

// matchesPrincipal checks if a policy applies to a user
func (e *PolicyEngine) matchesPrincipal(principals []string, user *models.User) bool {
	// Empty principals means applies to everyone
	if len(principals) == 0 {
		return true
	}

	for _, principal := range principals {
		// Check for wildcard
		if principal == "*" {
			return true
		}

		// Check for user ID
		if principal == user.ID {
			return true
		}

		// Check for username
		if principal == user.Username {
			return true
		}

		// Check for role (format: "role:rolename")
		if strings.HasPrefix(principal, "role:") {
			roleName := strings.TrimPrefix(principal, "role:")
			if e.hasRole(user, roleName) {
				return true
			}
		}

		// Check for group (format: "group:groupname")
		if strings.HasPrefix(principal, "group:") {
			groupName := strings.TrimPrefix(principal, "group:")
			if e.hasGroup(user, groupName) {
				return true
			}
		}
	}

	return false
}

// hasRole checks if a user has a specific role
func (e *PolicyEngine) hasRole(user *models.User, role string) bool {
	for _, r := range user.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// hasGroup checks if a user belongs to a specific group
func (e *PolicyEngine) hasGroup(user *models.User, group string) bool {
	for _, g := range user.Groups {
		if g == group {
			return true
		}
	}
	return false
}
