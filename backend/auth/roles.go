package auth

import (
	"github.com/andriyg76/bgl/models"
	"os"
	"strings"
)

// superAdminsCache caches the list of super admin IDs from environment
var superAdminsCache []string

func init() {
	loadSuperAdmins()
}

// loadSuperAdmins loads the super admin IDs from environment variable
func loadSuperAdmins() {
	superAdminsStr := os.Getenv("SUPERADMINS")
	if superAdminsStr == "" {
		superAdminsCache = []string{}
		return
	}
	superAdminsCache = strings.Split(superAdminsStr, ",")

	// Trim spaces
	for i := range superAdminsCache {
		superAdminsCache[i] = strings.TrimSpace(superAdminsCache[i])
	}
}

// IsSuperAdmin checks if the user is a super administrator
// It checks against the user's external IDs
func IsSuperAdmin(user *models.User) bool {
	if user == nil {
		return false
	}
	return IsSuperAdminByExternalIDs(user.ExternalIDs)
}

// IsSuperAdminByExternalIDs checks if any of the provided external IDs
// match a super admin ID
func IsSuperAdminByExternalIDs(externalIDs []string) bool {
	for _, adminID := range superAdminsCache {
		for _, userID := range externalIDs {
			if adminID == userID {
				return true
			}
		}
	}
	return false
}

// GetSuperAdmins returns the list of super admin IDs
// This is useful for testing or debugging
func GetSuperAdmins() []string {
	return superAdminsCache
}

// SetSuperAdminsForTesting allows tests to override the superadmins list
// Returns a function to restore the original value
func SetSuperAdminsForTesting(admins []string) func() {
	original := superAdminsCache
	superAdminsCache = admins
	return func() { superAdminsCache = original }
}
