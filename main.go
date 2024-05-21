package mirmiddleware

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

var URL string

func InitMiddleware(url string) {
	URL = url
}

type middleware struct {
	UserId int `json:"userId"`
}

func MiddlewareUser(c *fiber.Ctx) error {

	req, err := http.NewRequest("POST", URL, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 || resp.StatusCode == 401 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Forbidden",
		})
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var middlewareStruct middleware

	if err := json.Unmarshal([]byte(body), &middlewareStruct); err != nil {
		return err
	}

	c.Locals("userId", middlewareStruct.UserId)

	return c.Next()
}
