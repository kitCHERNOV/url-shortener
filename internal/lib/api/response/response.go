package response

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK                  = "OK"
	StatusError               = "Error"
	StatusBadRequestError     = "Bad Request Error"
	StatusInternalServerError = "Internal Server Error"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func BadRequestError(msg string) Response {
	return Response{
		Status: StatusBadRequestError,
		Error:  msg,
	}
}

func InternalServerError(msg string) Response {
	return Response{
		Status: StatusInternalServerError,
		Error:  msg,
	}
}

// TODO: validation error to response
