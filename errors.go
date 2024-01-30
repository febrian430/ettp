package ettp

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidMessage     = errors.New("invalid ettp message")
	ErrActionNotFound     = errors.New("action not found")
	ErrConnectionToServer = errors.New("failed to connect to tcp server")
)

func WrapErr(message string, err error) error {
	return fmt.Errorf("%s: %w", message, err)
}
