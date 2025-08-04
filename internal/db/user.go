package db

import (
	"database/sql"

	"github.com/samrityy/go-oauth/internal/models"
)

func SaveOrUpdateUser(db *sql.DB, user *models.User) (int, error) {
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE email = $1", user.Email).Scan(&userID)
	if err == sql.ErrNoRows {
		err = db.QueryRow(
			"INSERT INTO users (name, email, avatar_url) VALUES ($1, $2, $3) RETURNING id",
			user.Name, user.Email, user.AvatarURL,
		).Scan(&userID)
	}
	return userID, err
}
func SaveOrUpdateUserOAuth(db *sql.DB, oauth *models.UserOAuth) error {
	_, err := db.Exec(`
		INSERT INTO user_oauth (user_id, provider, provider_id, access_token, refresh_token)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (provider, provider_id)  
		DO UPDATE SET
			user_id = EXCLUDED.user_id,
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token
	`, oauth.UserID, oauth.Provider, oauth.ProviderID, oauth.AccessToken, oauth.RefreshToken)
	return err
}
