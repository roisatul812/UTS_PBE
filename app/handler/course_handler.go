package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"uts-pbe-siakad/app/model"
	"uts-pbe-siakad/app/service"
	"uts-pbe-siakad/helper"
)

type CourseHandler struct {
	courseService service.CourseService
}

func NewCourseHandler(courseService service.CourseService) *CourseHandler {
	return &CourseHandler{courseService: courseService}
}

// GET /api/v1/courses
func (h *CourseHandler) GetAll(c *fiber.Ctx) error {
	semester, _ := strconv.Atoi(c.Query("semester", "0"))
	availableStr := strings.ToLower(c.Query("available", "false"))
	available := (availableStr == "true" || availableStr == "1")

	query := model.CourseListQuery{
		Semester:  semester,
		Search:    c.Query("search"),
		Available: available,
	}

	courses, err := h.courseService.GetAll(query)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil daftar mata kuliah", nil)
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Daftar mata kuliah berhasil diambil", courses)
}
