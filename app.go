// app.go
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// UserInfo represents information retrieved from user APIs.
type UserInfo struct {
	ID        int    `json:"id"`
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
}

// Root renders the home page.
func (a *App) Root(w http.ResponseWriter, r *http.Request) {
	if a.AccessToken == "" {
		w.WriteHeader(http.StatusOK)
		if err := a.Template.Execute(w, a); err != nil {
			a.Logger.Error("failed executing template", "error", err)
		}
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		a.Logger.Error("failed creating request", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.AccessToken))
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		a.Logger.Error("failed retrieving user details", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		a.Logger.Error("failed decoding user details", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	emailReq, _ := http.NewRequest(http.MethodGet, "https://api.github.com/user/emails", nil)
	emailReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.AccessToken))
	emailReq.Header.Set("Accept", "application/vnd.github+json")

	emailResp, err := http.DefaultClient.Do(emailReq)
	if err == nil {
		defer emailResp.Body.Close()
		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := json.NewDecoder(emailResp.Body).Decode(&emails); err == nil {
			for _, e := range emails {
				if e.Primary && e.Verified {
					userInfo.Email = e.Email
					break
				}
			}
		}
	}

	a.UserInfo = &userInfo

	w.WriteHeader(http.StatusOK)
	if err := a.Template.Execute(w, a); err != nil {
		a.Logger.Error("failed executing template", "error", err)
		return
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
	req.Header.Set("User-Agent", "YourAppName")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
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

	return &user, nil
}

func getFacebookUserInfo(accessToken string) (*UserInfo, error) {
	resp, err := http.Get("https://graph.facebook.com/me?fields=id,name,email,picture&access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var fbData struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&fbData); err != nil {
		return nil, err
	}

	return &UserInfo{
		Login:     fbData.Name,
		Email:     fbData.Email,
		AvatarURL: fbData.Picture.Data.URL,
	}, nil
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
