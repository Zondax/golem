package domain

import "fmt"

type APIError struct {
	HTTPStatus int    `json:"-"`
	ErrorCode  string `json:"error_code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
}

func (ae *APIError) Error() string {
	return fmt.Sprintf("HTTP Status: %d, ErrorCode: %s, Message: %s", ae.HTTPStatus, ae.ErrorCode, ae.Message)
}

func NewAPIErrorResponse(httpStatus int, errorCode, message string, details ...string) *APIError {
	apiError := &APIError{
		HTTPStatus: httpStatus,
		ErrorCode:  errorCode,
		Message:    message,
	}

	if len(details) > 0 {
		apiError.Details = details[0]
	}

	return apiError
}
