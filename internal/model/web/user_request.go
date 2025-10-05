package web

type RegisterUserRequest struct {
	Email    string `validate:"required,email,max=255"`
	FullName string `validate:"required,min=3,max=255,validName"`
	Password string `validate:"required,min=6,max=255"`
	Role     string `validate:"required,oneof=student teacher"`
}
