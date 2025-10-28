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

type EssayCorrection struct {
	StudentAnswerId string  `json:"student_answer_id"`
	Question        string  `json:"question"`
	StudentAnswer   string  `json:"student_answer"`
	Score           float64 `json:"score"`
	Feedback        string  `json:"feedback"`
	MaxScore        float64 `json:"max_score"`
	Similarity      float64 `json:"similarity"`
}
