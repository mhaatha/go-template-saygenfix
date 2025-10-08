package domain

import "time"

type QAItem struct {
	Id       string
	Question string `json:"question"`
	Answer   string `json:"answer"`
	ExamId   string
}

type Exam struct {
	Id        string
	RoomName  string
	Year      int
	Duration  int
	TeacherId string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
