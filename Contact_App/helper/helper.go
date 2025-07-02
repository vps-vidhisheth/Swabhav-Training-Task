package helper

// UserData is used for static helper checks
type UserData struct {
	IsAdmin  bool
	IsActive bool
}

// Authorizer defines an interface for checking user roles
type Authorizer interface {
	IsAdminActive() bool
	IsStaffActive() bool
}

// Static check functions (can still be useful)
func IsAuthorizedAdmin(user UserData) bool {
	return user.IsAdmin && user.IsActive
}

func IsAuthorizedStaff(user UserData) bool {
	return !user.IsAdmin && user.IsActive
}
