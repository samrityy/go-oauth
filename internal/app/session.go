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

func GetAuthCookies(r *http.Request) (string, string, error) {
	accessTokenCookie, err := r.Cookie("access_token")
	if err != nil {
		return "", "", err
	}

	userIDCookie, err := r.Cookie("user_id")
	if err != nil {
		return "", "", err
	}

	return accessTokenCookie.Value, userIDCookie.Value, nil
}


func ClearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:    "user_id",
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})
}
