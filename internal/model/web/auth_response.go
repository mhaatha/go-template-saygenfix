package web

import "time"

type LoginResponse struct {
	Id        int
	Email     string
	FullName  string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
