package validator

import (
	"strings"
	"unicode"

	"loyalty/internal/domain/order/model"
	"loyalty/internal/domain/order/service"
)

// Validator реализует service.OrderNumberValidator.
type Validator struct{}

// NewValidator создаёт валидатор номеров заказов.
func NewValidator() *Validator { return &Validator{} }

// ValidateNumber нормализует номер заказа и проверяет его корректность.
// Номер должен состоять только из цифр и проходить проверку алгоритмом Луна.
func (v *Validator) ValidateNumber(number string) (string, error) {
	normalized := strings.TrimSpace(number)
	if normalized == "" {
		return "", model.ErrInvalidOrderNumber
	}
	for _, digit := range normalized {
		if !unicode.IsDigit(digit) {
			return "", model.ErrInvalidOrderNumber
		}
	}
	if !luhnValid(normalized) {
		return "", model.ErrInvalidOrderNumber
	}
	return normalized, nil
}

func luhnValid(number string) bool {
	var sum int
	var alt bool
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if alt {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alt = !alt
	}
	return sum%10 == 0
}

var _ service.OrderNumberValidator = (*Validator)(nil)
