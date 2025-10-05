package web

type LoginRequest struct {
	Email    string `validate:"required,email,max=255"`
	Password string `validate:"required,min=6,max=255"`
}
