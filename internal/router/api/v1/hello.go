package v1

import (
	"chat_backend/internal/response"

	"github.com/labstack/echo/v4"
)

func Hello(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		name = "World"
	}

	data := map[string]string{
		"message": "Hello, " + name,
	}

	return response.Success(c, data)
}
