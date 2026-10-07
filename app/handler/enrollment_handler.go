package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"uts-pbe-siakad/app/model"
	"uts-pbe-siakad/app/service"
	"uts-pbe-siakad/helper"
)

type EnrollmentHandler struct {
	enrollmentService service.EnrollmentService
}

func NewEnrollmentHandler(enrollmentService service.EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{enrollmentService: enrollmentService}
}

// POST /api/v1/enrollments
func (h *EnrollmentHandler) Create(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int)
	if !ok {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Autentikasi diperlukan", nil)
	}

	var req model.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Format request body tidak valid", map[string][]string{
			"body": {"Request body harus berupa JSON yang valid"},
		})
	}

	data, statusCode, message, fieldErrors := h.enrollmentService.Enroll(userID, req)
	if statusCode != fiber.StatusCreated {
		return helper.ErrorResponse(c, statusCode, message, fieldErrors)
	}

	return helper.SuccessResponse(c, fiber.StatusCreated, message, data)
}

// DELETE /api/v1/enrollments/:id
func (h *EnrollmentHandler) Delete(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int)
	if !ok {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Autentikasi diperlukan", nil)
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Data enrollment tidak ditemukan", nil)
	}

	statusCode, message, serviceErr := h.enrollmentService.Unenroll(userID, id)
	if statusCode != fiber.StatusNoContent {
		return helper.ErrorResponse(c, statusCode, message, nil)
	}
	if serviceErr != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Terjadi kesalahan internal server", nil)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
