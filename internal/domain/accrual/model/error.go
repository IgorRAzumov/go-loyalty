package model

import "errors"

// ErrTooManyRequests возвращается при превышении rate limit (429 Too Many Requests).
var ErrTooManyRequests = errors.New("accrual system rate limit exceeded")
