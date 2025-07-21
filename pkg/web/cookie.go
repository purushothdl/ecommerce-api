// pkg/web/cookie.go
package web

import (
	"net/http"
)

// CookieSettings defines the parameters for setting a cookie.
type CookieSettings struct {
	Name       string
	Value      string
	Path       string
	MaxAge     int // in seconds
	IsHTTPOnly bool
	IsSecure   bool
	SameSite   http.SameSite
}

// SetCookie is a generic helper to set a cookie on an http.ResponseWriter.
func SetCookie(w http.ResponseWriter, settings CookieSettings) {
	http.SetCookie(w, &http.Cookie{
		Name:     settings.Name,
		Value:    settings.Value,
		Path:     settings.Path,
		MaxAge:   settings.MaxAge,
		HttpOnly: settings.IsHTTPOnly,
		Secure:   settings.IsSecure,
		SameSite: settings.SameSite,
	})
}

// ClearCookie clears a cookie by setting its MaxAge to -1.
func ClearCookie(w http.ResponseWriter, name string, isProduction bool) {
	SetCookie(w, CookieSettings{
		Name:       name,
		Value:      "",
		Path:       "/",
		MaxAge:     -1,
		IsHTTPOnly: true,
		IsSecure:   isProduction,
		SameSite:   http.SameSiteStrictMode,
	})
}

// Example usage for Refresh Token
const RefreshTokenCookieName = "refresh_token"

func SetRefreshTokenCookie(w http.ResponseWriter, token string, isProduction bool) {
	SetCookie(w, CookieSettings{
		Name:       RefreshTokenCookieName,
		Value:      token,
		Path:       "/",
		MaxAge:     7 * 24 * 60 * 60, // 7 days
		IsHTTPOnly: true,
		IsSecure:   isProduction,
		SameSite:   http.SameSiteStrictMode,
	})
}

// Example usage for Cart Token
const CartIDCookieName = "cart_id"

func SetCartCookie(w http.ResponseWriter, cartID string, isProduction bool) {
	SetCookie(w, CookieSettings{
		Name:       CartIDCookieName,
		Value:      cartID,
		Path:       "/",
		MaxAge:     30 * 24 * 60 * 60, // 30 days
		IsHTTPOnly: true,
		IsSecure:   isProduction,
		SameSite:   http.SameSiteLaxMode, // Lax is often better for carts
	})
}