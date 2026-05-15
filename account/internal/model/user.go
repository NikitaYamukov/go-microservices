package model

import "time"

type User struct {
	ID         uint64
	Login      string
	Email      string
	Phone      string
	FirstName  string
	LastName   string
	MiddleName string
	Age        uint32
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type CreateUser struct {
	Login      string
	Email      string
	Phone      string
	FirstName  string
	LastName   string
	MiddleName string
	Age        uint32
}

type UpdateUser struct {
	Email      string
	Phone      string
	FirstName  string
	LastName   string
	MiddleName string
	Age        uint32
}
