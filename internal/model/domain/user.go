package domain

import "time"

type User struct {
	Id        string
	Email     string
	FullName  string
	Password  string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
