package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword хеширует пароль с использованием bcrypt.
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// ComparePassword сравнивает bcrypt-хеш с введённым паролем.
func ComparePassword(hash []byte, password string) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(password))
}
