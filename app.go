// app.go
package main

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// UserInfo represents information retrieved from user APIs.
type UserInfo struct {
	ID        string `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// App holds configuration and in-memory session data.
type App struct {
	OAuthConfigs map[string]*oauth2.Config
	Logger       *slog.Logger
	Template     *template.Template

	AccessToken  string
	RefreshToken string
	UserInfo     *UserInfo
	Provider     string
}

// Root renders the home page.
func (a *App) Root(w http.ResponseWriter, r *http.Request) {
	if a.AccessToken == "" || a.UserInfo == nil {
		w.WriteHeader(http.StatusOK)
		_ = a.Template.Execute(w, a)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := a.Template.Execute(w, a); err != nil {
		a.Logger.Error("failed executing template", "error", err)
	}
}

// OAuthCallback handles OAuth2 callback for any provider.
func (a *App) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := strings.TrimPrefix(r.URL.Path, "/oauth2/callback/")
	oauthCfg, ok := a.OAuthConfigs[provider]
	if !ok {
		http.Error(w, "Unknown provider", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := oauthCfg.Exchange(r.Context(), code)
	if err != nil {
		a.Logger.Error("failed oauth exchange", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	a.Logger.Info("completed oauth exchange",
		"token_type", token.Type(),
		"expiration", token.Expiry,
	)

	a.AccessToken = token.AccessToken
	a.RefreshToken = token.RefreshToken

	var userInfo *UserInfo
	switch provider {
	case "github":
		userInfo, err = getGitHubUserInfo(a.AccessToken)
	case "facebook":
		userInfo, err = getFacebookUserInfo(a.AccessToken)
	case "google":
		userInfo, err = getGoogleUserInfo(a.AccessToken)
	default:
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	if err != nil {
		a.Logger.Error("failed retrieving user info", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	a.UserInfo = userInfo
	a.Provider = provider

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// LoggerMiddleware logs the start and end of requests.
func LoggerMiddleware(logger *slog.Logger, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		receivedTime := time.Now()
		logger.Info("request received", "method", r.Method, "path", r.URL.Path)
		handler(w, r)
		logger.Info("request complete", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(receivedTime).Milliseconds())
	}
}

func getGitHubUserInfo(accessToken string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Oauth")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ghData struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ghData); err != nil {
		return nil, err
	}

	user := &UserInfo{
		ID:        strconv.FormatInt(ghData.ID, 10),
		Login:     ghData.Login,
		Name:      ghData.Name,
		AvatarURL: ghData.AvatarURL,
	}

	req2, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return nil, err
	}
	req2.Header.Set("Authorization", "Bearer "+accessToken)
	req2.Header.Set("Accept", "application/vnd.github+json")
	req2.Header.Set("User-Agent", "YourAppName")

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&emails); err != nil {
		return nil, err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			user.Email = e.Email
			break
		}
	}

	return user, nil
}

func getFacebookUserInfo(accessToken string) (*UserInfo, error) {
	resp, err := http.Get("https://graph.facebook.com/me?fields=id,name,email,picture&access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var fbData struct {
		ID      string  `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&fbData); err != nil {
		return nil, err
	}

	user := &UserInfo{
		ID:        fbData.ID,
		Name:      fbData.Name,
		Email:     fbData.Email,
		AvatarURL: fbData.Picture.Data.URL,
	}

	return user, nil
}


func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	provider := strings.TrimPrefix(r.URL.Path, "/login/")
	oauthCfg, ok := a.OAuthConfigs[provider]
	if !ok {
		http.Error(w, "Unknown provider", http.StatusBadRequest)
		return
	}

	url := oauthCfg.AuthCodeURL("state-" + provider)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
