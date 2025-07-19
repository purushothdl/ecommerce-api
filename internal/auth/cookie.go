package auth

import (
	"net/http"
)

const refreshTokenCookieName = "refresh_token"

// SetRefreshTokenCookie sets the refresh token in a secure, http-only cookie.
func SetRefreshTokenCookie(w http.ResponseWriter, token string, isProduction bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		HttpOnly: true,
		Secure:   isProduction, // Should be true in production
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days in seconds
	})
}

// ClearRefreshTokenCookie expires the refresh token cookie, effectively deleting it.
func ClearRefreshTokenCookie(w http.ResponseWriter, isProduction bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   isProduction, // Should be true in production
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1, // A value < 0 instructs the browser to delete the cookie
	})
}