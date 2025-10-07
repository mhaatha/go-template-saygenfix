package domain

type QAItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type Exam struct {
	RoomName string
	Year     int
	Duration int
}
