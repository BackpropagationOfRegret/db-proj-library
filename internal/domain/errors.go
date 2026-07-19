package domain

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrCopyUnavailable    = errors.New("copy unavailable")
	ErrReaderBlocked      = errors.New("reader blocked")
	ErrLoanLimitExceeded  = errors.New("loan limit exceeded")
	ErrLoanNotActive      = errors.New("loan is not active")
	ErrAlreadyExists      = errors.New("already exists")
	ErrInvalidArgument    = errors.New("invalid argument")
)
