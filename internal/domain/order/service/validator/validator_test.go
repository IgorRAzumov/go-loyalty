package validator

import (
	"testing"
)

func TestValidator_ValidateNumber(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid luhn",
			input:   "12345678903",
			want:    "12345678903",
			wantErr: false,
		},
		{
			name:    "valid luhn 2",
			input:   "79927398713",
			want:    "79927398713",
			wantErr: false,
		},
		{
			name:    "invalid luhn",
			input:   "12345678901",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "non-digit characters",
			input:   "abc123",
			want:    "",
			wantErr: true,
		},
		{
			name:    "single digit valid",
			input:   "0",
			want:    "0",
			wantErr: false,
		},
		{
			name:    "valid with leading zeros",
			input:   "00",
			want:    "00",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := v.ValidateNumber(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}
