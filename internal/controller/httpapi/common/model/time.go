package model

import "time"

// RFC3339Time сериализует время строго в формате RFC3339 (без наносекунд).
type RFC3339Time struct {
	time.Time
}

// MarshalJSON сериализует время в JSON-строку в формате RFC3339.
func (t RFC3339Time) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	s := t.Time.Format(time.RFC3339)
	return []byte(`"` + s + `"`), nil
}
