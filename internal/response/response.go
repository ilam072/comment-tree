package response

import (
	"github.com/wb-go/wbf/ginext"
)

// Response — стандартная структура для JSON-ответа.
type Response struct {
	Status  string      `json:"status"`
	Payload interface{} `json:"payload,omitempty"`
}

// Success создаёт успешный Response с данными.
func Success(payload interface{}) Response {
	return Response{
		Status:  "ok",
		Payload: payload,
	}
}

// Error создаёт Response с ошибкой.
func Error(payload interface{}) Response {
	return Response{
		Status:  "error",
		Payload: payload,
	}
}

// WriteJSON отправляет Response через Gin с указанным HTTP кодом.
func (r Response) WriteJSON(c *ginext.Context, code int) {
	c.JSON(code, r)
}

func Raw(c *ginext.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}
