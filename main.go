package mirmiddleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

var URL string

func InitMiddleware(url string) {
	URL = url
}

type Role struct {
	ID          int     `json:"id"`
	Label       string  `json:"label"`
	AccessLevel int     `json:"accessLevel"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
	DeletedAt   *string `json:"deletedAt"`
}

type middleware struct {
	UserId int    `json:"userId"`
	Roles  []Role `json:"roles"`
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

func MiddlewareRole(roles []int, strict bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return middlewareRoleWrapper(roles, strict, c)
	}
}

func middlewareRoleWrapper(roles []int, strict bool, c *fiber.Ctx) error {
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
		fmt.Printf("Failed to unmarshal: %v\nBody: %s\n", err, string(body))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "invalid server response",
			"error":   err.Error(),
		})
	}

	// Создаем типа объекта для хранения id ролей пользователя
	userRoleIDs := make(map[int]bool)
	for _, role := range middlewareStruct.Roles {
		userRoleIDs[role.ID] = true
	}

	// Проверка ролей
	hasRequiredRoles := false
	if strict {
		// Строгий режим,  все указанные роли должны быть
		hasRequiredRoles = true
		for _, roleId := range roles {
			if !userRoleIDs[roleId] {
				hasRequiredRoles = false
				break
			}
		}
	} else {
		// Не строгий режим, хотя бы одна из указанных ролей должна быть
		for _, roleId := range roles {
			if userRoleIDs[roleId] {
				hasRequiredRoles = true
				break
			}
		}
	}

	if !hasRequiredRoles {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Forbidden",
		})
	}

	c.Locals("userId", middlewareStruct.UserId)

	return c.Next()
}
