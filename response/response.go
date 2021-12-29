package response

import "github.com/labstack/echo/v4"

type Dto struct {
	Status  int         `json:"status"`
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type Inf interface {
	SuccessResponse(e echo.Context, status int, code string, message string)
	ErrorResponse(e echo.Context, status int, code string, message string, data interface{})
}

type response struct{}

func Helper() Inf {
	return new(response)
}

func (t *response) SuccessResponse(e echo.Context, status int, code string, message string) {
	responseData := Dto{
		Status:  status,
		Code:    code,
		Message: message,
	}
	_ = e.JSON(status, responseData)
	return
}
func (t *response) ErrorResponse(e echo.Context, status int, code string, message string, data interface{}) {
	responseData := Dto{
		Status:  status,
		Code:    code,
		Message: message,
		Data:    data,
	}
	_ = e.JSON(status, responseData)
	return
}
