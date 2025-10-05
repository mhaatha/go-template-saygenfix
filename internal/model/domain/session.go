package domain

import "time"

type Session struct {
	SessionId string
	UserId    string
	CreatedAt time.Time
}
