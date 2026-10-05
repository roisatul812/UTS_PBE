package handler

import (
	"github.com/gofiber/fiber/v2"
	"uts-pbe-siakad/app/model"
	"uts-pbe-siakad/app/service"
	"uts-pbe-siakad/helper"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Format request body tidak valid", map[string][]string{
			"body": {"Request body harus berupa JSON yang valid"},
		})
	}

	clientIP := c.IP()
	data, statusCode, message, fieldErrors := h.authService.Login(req, clientIP)
	if statusCode != fiber.StatusOK {
		return helper.ErrorResponse(c, statusCode, message, fieldErrors)
	}

	return helper.SuccessResponse(c, fiber.StatusOK, message, data)
}

// GET /api/v1/auth/me
func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int)
	if !ok {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Autentikasi diperlukan", nil)
	}

	data, statusCode, message, err := h.authService.GetMe(userID)
	if statusCode != fiber.StatusOK {
		return helper.ErrorResponse(c, statusCode, message, nil)
	}
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Terjadi kesalahan internal server", nil)
	}

	return helper.SuccessResponse(c, fiber.StatusOK, message, data)
}
