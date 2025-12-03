package server

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/Siroshun09/serrors"
	"github.com/okocraft/auth-service/internal/domain"
)

func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", serrors.WithStackTrace(err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func checkCSRFToken(r *http.Request, csrfToken *string) error {
	csrfTokenCookie, err := r.Cookie("csrf_token")
	if err != nil {
		return domain.NewUnauthorizedError(serrors.New("csrf token not found"))
	}

	if csrfToken == nil || *csrfToken != csrfTokenCookie.Value {
		return domain.NewUnauthorizedError(serrors.New("csrf token mismatch"))
	}

	return nil
}

func setRefreshTokenCookie(w http.ResponseWriter, refreshToken string, csrfToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  expiresAt,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: false,
		Secure:   true,
		Expires:  expiresAt,
		SameSite: http.SameSiteLaxMode,
	})
}

func unsetRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}
