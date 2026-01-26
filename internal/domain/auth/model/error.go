package model

import "errors"

var (
	// ErrLoginTaken возвращается при попытке зарегистрировать уже занятый логин.
	ErrLoginTaken = errors.New("login already taken")
	// ErrInvalidCreds возвращается при неверной паре логин/пароль.
	ErrInvalidCreds = errors.New("invalid credentials")
	// ErrInvalidInput возвращается при нарушении базовых ограничений на входные данные.
	ErrInvalidInput = errors.New("invalid input")
	// ErrPasswordTooShort возвращается, если пароль не удовлетворяет минимальным требованиям.
	ErrPasswordTooShort = errors.New("password too short")
	// ErrPasswordTooLong возвращается, если пароль превышает допустимую длину.
	ErrPasswordTooLong = errors.New("password too long")
	// ErrInvalidToken возвращается при невалидном/истёкшем токене.
	ErrInvalidToken = errors.New("invalid token")

	// ErrNotFound возвращается, когда сущность не найдена (например, пользователь по логину).
	ErrNotFound = errors.New("not found")
)
