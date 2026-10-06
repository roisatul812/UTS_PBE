package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"uts-pbe-siakad/app/model"
	"uts-pbe-siakad/app/service"
	"uts-pbe-siakad/helper"
)

type StudentHandler struct {
	studentService service.StudentService
}

func NewStudentHandler(studentService service.StudentService) *StudentHandler {
	return &StudentHandler{studentService: studentService}
}

// GET /api/v1/students
func (h *StudentHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	angkatan, _ := strconv.Atoi(c.Query("angkatan", "0"))

	query := model.StudentListQuery{
		Page:     page,
		PerPage:  perPage,
		Prodi:    c.Query("prodi"),
		Angkatan: angkatan,
		Search:   c.Query("search"),
		Sort:     c.Query("sort"),
	}

	students, meta, err := h.studentService.GetAll(query)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data mahasiswa", nil)
	}

	return helper.PaginatedResponse(c, fiber.StatusOK, "Data mahasiswa berhasil diambil", students, meta)
}

// POST /api/v1/students
func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Format request body tidak valid", map[string][]string{
			"body": {"Request body harus berupa JSON yang valid"},
		})
	}

	student, statusCode, message, fieldErrors := h.studentService.Create(req)
	if statusCode != fiber.StatusCreated {
		return helper.ErrorResponse(c, statusCode, message, fieldErrors)
	}

	return helper.SuccessResponse(c, fiber.StatusCreated, message, student)
}

// GET /api/v1/students/:id
func (h *StudentHandler) GetDetail(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Data mahasiswa tidak ditemukan", nil)
	}

	currentUserID := c.Locals("user_id").(int)
	currentUserRole := c.Locals("role").(string)

	detail, statusCode, message, serviceErr := h.studentService.GetDetail(id, currentUserID, currentUserRole)
	if statusCode != fiber.StatusOK {
		return helper.ErrorResponse(c, statusCode, message, nil)
	}
	if serviceErr != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Terjadi kesalahan internal server", nil)
	}

	return helper.SuccessResponse(c, fiber.StatusOK, message, detail)
}

// PUT /api/v1/students/:id
func (h *StudentHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Data mahasiswa tidak ditemukan", nil)
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusUnprocessableEntity, "Format request body tidak valid", map[string][]string{
			"body": {"Request body harus berupa JSON yang valid"},
		})
	}

	updated, statusCode, message, fieldErrors := h.studentService.Update(id, req)
	if statusCode != fiber.StatusOK {
		return helper.ErrorResponse(c, statusCode, message, fieldErrors)
	}

	return helper.SuccessResponse(c, fiber.StatusOK, message, updated)
}

// DELETE /api/v1/students/:id
func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return helper.ErrorResponse(c, fiber.StatusNotFound, "Data mahasiswa tidak ditemukan", nil)
	}

	statusCode, message, serviceErr := h.studentService.Delete(id)
	if statusCode != fiber.StatusNoContent {
		return helper.ErrorResponse(c, statusCode, message, nil)
	}
	if serviceErr != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, "Terjadi kesalahan internal server", nil)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
