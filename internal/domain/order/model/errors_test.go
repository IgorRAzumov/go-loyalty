package model

import (
	"errors"
	"testing"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrInvalidOrderNumber", ErrInvalidOrderNumber},
		{"ErrOrderAlreadyUploaded", ErrOrderAlreadyUploaded},
		{"ErrOrderAlreadyUploadedByAnother", ErrOrderAlreadyUploadedByAnother},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s is nil", tt.name)
			}
			if !errors.Is(tt.err, tt.err) {
				t.Errorf("errors.Is failed for %s", tt.name)
			}
		})
	}
}

func TestStatus(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   string
	}{
		{"NEW", StatusNew, "NEW"},
		{"PROCESSING", StatusProcessing, "PROCESSING"},
		{"INVALID", StatusInvalid, "INVALID"},
		{"PROCESSED", StatusProcessed, "PROCESSED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("Status = %v, want %v", tt.status, tt.want)
			}
		})
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrInvalidOrderNumber", ErrInvalidOrderNumber, "invalid order number"},
		{"ErrOrderAlreadyUploaded", ErrOrderAlreadyUploaded, "order already uploaded by this user"},
		{"ErrOrderAlreadyUploadedByAnother", ErrOrderAlreadyUploadedByAnother, "order already uploaded by another user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("Error() = %v, want %v", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestRateLimitError(t *testing.T) {
	tests := []struct {
		name       string
		err        RateLimitError
		wantPrefix string
	}{
		{"with retry", RateLimitError{RetryAfter: 60}, "accrual rate limited (retry after"},
		{"without retry", RateLimitError{}, "accrual rate limited"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if msg == "" {
				t.Error("Error() returned empty string")
			}
			if tt.err.Unwrap() != ErrAccrualRateLimited {
				t.Error("Unwrap() should return ErrAccrualRateLimited")
			}
		})
	}
}
