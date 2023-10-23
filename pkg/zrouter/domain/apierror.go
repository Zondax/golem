package domain

type APIError struct {
	HTTPStatus int    `json:"-"`
	ErrorCode  string `json:"error_code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
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
