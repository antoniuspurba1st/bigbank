package model

type APIResponse struct {
	Status        string      `json:"status"`
	Message       string      `json:"message"`
	CorrelationID string      `json:"correlation_id"`
	Data          interface{} `json:"data,omitempty"`
}

type APIError struct {
	Status        string `json:"status"`
	Code          string `json:"code"`
	Message       string `json:"message"`
	CorrelationID string `json:"correlation_id"`
}

type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Err        error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}

	return e.Message
}
