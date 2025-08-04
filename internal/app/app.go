// app.go
package app

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/samrityy/go-oauth/internal/db"
	"github.com/samrityy/go-oauth/internal/models"
	"golang.org/x/oauth2"
)

// UserInfo represents information retrieved from user APIs.

// App holds configuration and in-memory session data.
type App struct {
	OAuthConfigs map[string]*oauth2.Config
	Logger       *slog.Logger
	Template     *template.Template

	AccessToken  string
	RefreshToken string
	UserInfo     *models.User
	Provider     string
	DB           *sql.DB
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

func (a *App) SetupRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// OAuth & home routes
	mux.HandleFunc("/", a.Root)
	mux.HandleFunc("/login/", a.Login)
	mux.HandleFunc("/logout/", a.Logout)
	mux.HandleFunc("/oauth2/callback/", a.OAuthCallback)

	// User routes
	mux.HandleFunc("/users", a.HandleUsers)
	mux.HandleFunc("/users/", a.HandleUserByID)

	return mux
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

	var userInfo *models.User
	switch provider {
	case "github":
		userInfo, err = getGitHubUserInfo(a.AccessToken)
	case "facebook":
		userInfo, err = getFacebookUserInfo(a.AccessToken)
	case "google":
		userInfo, err = getGoogleUserInfo(a.AccessToken)
	case "instagram":
		userInfo, err = getInstagramUserInfo(os.Getenv("INSTAGRAM_ACCESS_TOKEN"))

	default:
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

	if err != nil {
		a.Logger.Error("failed retrieving user info", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userID, err := db.SaveOrUpdateUser(a.DB, userInfo)
	if err != nil {
		a.Logger.Error("failed to save user", "error", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	// Save OAuth tokens
	err = db.SaveOrUpdateUserOAuth(a.DB, &models.UserOAuth{
		UserID:       userID,
		Provider:     provider,
		ProviderID:   userInfo.ID,
		AccessToken:  a.AccessToken,
		RefreshToken: a.RefreshToken,
	})
	if err != nil {
		a.Logger.Error("failed to save oauth info", "error", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	SetAuthCookies(a.AccessToken, strconv.Itoa(userID), w)
	a.UserInfo = userInfo
	a.Provider = provider

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func getGitHubUserInfo(accessToken string) (*models.User, error) {
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
	user := &models.User{
		ID:        strconv.FormatInt(ghData.ID, 10),
		Name:      ghData.Login,
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

func getFacebookUserInfo(accessToken string) (*models.User, error) {
	resp, err := http.Get("https://graph.facebook.com/me?fields=id,name,email,picture&access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var fbData struct {
		ID      string `json:"id"`
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

	user := &models.User{
		ID:        fbData.ID,
		Name:      fbData.Name,
		Email:     fbData.Email,
		AvatarURL: fbData.Picture.Data.URL,
	}

	return user, nil
}
func getGoogleUserInfo(accessToken string) (*models.User, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var googleData struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleData); err != nil {
		return nil, err
	}

	user := &models.User{
		ID:        googleData.ID,
		Name:      googleData.Name,
		Email:     googleData.Email,
		AvatarURL: googleData.Picture,
	}

	return user, nil
}

func getInstagramUserInfo(accessToken string) (*models.User, error) {
	resp, err := http.Get("https://graph.instagram.com/me?fields=id,username,account_type&access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var igData struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&igData); err != nil {
		return nil, err
	}
	user := &models.User{
		ID:   igData.ID,
		Name: igData.Username,
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

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	a.AccessToken = ""
	a.RefreshToken = ""
	a.UserInfo = nil
	a.Provider = ""
	ClearAuthCookies(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
