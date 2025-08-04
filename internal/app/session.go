package app

import (
	"net/http"
	"time"
)

func SetAuthCookies(token string, userID string, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:    "user_id",
		Value:   userID,
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour),
	})
}

func GetSessionToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
