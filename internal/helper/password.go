package helper

import "golang.org/x/crypto/bcrypt"

func CheckPasswordConfirmation(password string, confirmPassword string) bool {
	return password == confirmPassword
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPasswordHash checks if the input password matches the hashed password
func CheckPasswordHash(hashedPassword, inputPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(inputPassword))
	return err == nil
}
