package helper

import (
	"github.com/gofiber/fiber/v2"
	"uts-pbe-siakad/app/model"
)

func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(model.ApiResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func PaginatedResponse(c *fiber.Ctx, statusCode int, message string, data interface{}, meta *model.PaginationMeta) error {
	return c.Status(statusCode).JSON(model.ApiResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func ErrorResponse(c *fiber.Ctx, statusCode int, message string, errors map[string][]string) error {
	return c.Status(statusCode).JSON(model.ApiResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}
