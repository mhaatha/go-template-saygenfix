package web

import "time"

type RegisterUserResponse struct {
	Id        int
	Email     string
	FullName  string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
