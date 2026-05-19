package model

import "time"

type User struct {
	ID           uint64
	Login        string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Register struct {
	Login    string
	Email    string
	Password string
}

type Login struct {
	LoginOrEmail string
	Password     string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type RefreshToken struct {
	ID        uint64
	UserID    uint64
	Token     string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}
