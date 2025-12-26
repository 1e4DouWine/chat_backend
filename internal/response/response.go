package response

import (
	"chat_backend/internal/errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`    // 0 表示成功，非 0 表示错误
	Message string      `json:"message"` // 错误描述或成功信息
	Data    interface{} `json:"data"`    // 响应数据，可能为 null
}

// Success 成功响应
func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, Response{
		Code:    errors.Success,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(c echo.Context, code int, message string) error {
	return c.JSON(getHTTPStatus(code), Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c echo.Context, code int, message string, data interface{}) error {
	return c.JSON(getHTTPStatus(code), Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// getHTTPStatus 根据业务错误码获取HTTP状态码
func getHTTPStatus(code int) int {
	switch {
	case code >= 1000 && code < 2000:
		return http.StatusBadRequest
	case code >= 2000 && code < 3000:
		return http.StatusUnauthorized
	case code >= 3000 && code < 4000:
		return http.StatusForbidden
	case code >= 4000 && code < 5000:
		return http.StatusNotFound
	case code >= 5000 && code < 6000:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
