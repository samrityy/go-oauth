package models


import "time"

type User struct {
    ID           string   
    Provider     string    
    ProviderID   string    
    Login        string
    Name         string
    Email        string
    AvatarURL    string
    AccessToken  string
    RefreshToken string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}