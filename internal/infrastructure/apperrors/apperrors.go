package apperrors

import (
	"fmt"
	"net/http"
)

const (
	BadRequest Type = "BAD_REQUEST"
	Internal   Type = "INTERNAL"
)

type (
	Type string

	AppError struct {
		Type    Type
		Message map[string]string
	}
)

func New(Type Type, msg string) *AppError {
	return &AppError{
		Type:    Type,
		Message: map[string]string{"error": msg},
	}
}

func (err *AppError) Error() string {
	return err.Message["error"]
}

func (err *AppError) Status() int {
	switch err.Type {
	case BadRequest:
		return http.StatusBadRequest
	case Internal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

var (
	ErrDatabase   = New(Internal, "Database raised an error")
	ErrInternal   = New(Internal, "Internal server error")
	ErrBadRequest = New(BadRequest, "Bad request")
)

func ErrDatabaseMsg(msg string) string {
	return fmt.Sprintf("Database raised an error: %s", msg)
}

func ErrInternalMsg(msg string) string {

	return fmt.Sprintf("Internal server error: %s", msg)

}

func ErrBadRequestMsg(msg string) string {

	return fmt.Sprintf("Bad request: %s", msg)

}

func ErrBadRequestf(msg string) *AppError {

	return New(BadRequest, fmt.Sprintf("Bad Request: %s", msg))

}
