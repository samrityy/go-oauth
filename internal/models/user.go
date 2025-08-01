package models


import "time"

type User struct {
    ID           string       
    Name         string
    Email        string
    AvatarURL    string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}