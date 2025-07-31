package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/samrityy/go-oauth/internal/models"
)

var db *sql.DB

func GetUser(w http.ResponseWriter, r *http.Request) {
	//checks id this is a Get request
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Error reading users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
