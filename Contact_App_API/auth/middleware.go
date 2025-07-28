package auth

import (
	"Contact_App/helper"
	"net/http"
)

func MiddlewareAdminActive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserClaims(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userData := helper.UserData{
			IsAdmin:  claims.IsAdmin,
			IsActive: claims.IsActive,
		}

		if !helper.IsAuthorizedAdmin(userData) {
			http.Error(w, "Forbidden: admin privileges required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func MiddlewareStaffActive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserClaims(r)
		if claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userData := helper.UserData{
			IsAdmin:  claims.IsAdmin,
			IsActive: claims.IsActive,
		}

		if !helper.IsAuthorizedStaff(userData) {
			http.Error(w, "Forbidden: staff privileges required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ✅ New Middleware: Allow both active staff and active admin
func MiddlewareUserActive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserClaims(r)
		if claims == nil || !claims.IsActive {
			http.Error(w, "Unauthorized or inactive user", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
