package route

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"uts-pbe-siakad/app/handler"
	"uts-pbe-siakad/middleware"
)

type RouterConfig struct {
	App               *fiber.App
	DB                *sql.DB
	AuthHandler       *handler.AuthHandler
	StudentHandler    *handler.StudentHandler
	CourseHandler     *handler.CourseHandler
	EnrollmentHandler *handler.EnrollmentHandler
}

func SetupRoutes(cfg *RouterConfig) {
	api := cfg.App.Group("/api/v1")

	// 1. Auth routes (Endpoints 1 & 2)
	auth := api.Group("/auth")
	auth.Post("/login", cfg.AuthHandler.Login)
	auth.Get("/me", middleware.AuthMiddleware(cfg.DB), cfg.AuthHandler.GetMe)

	// 2. Student routes (Endpoints 3, 4, 5, 6, 7)
	if cfg.StudentHandler != nil {
		students := api.Group("/students")
		students.Use(middleware.AuthMiddleware(cfg.DB))

		// Endpoint 3: GET /api/v1/students (Admin only)
		students.Get("/", middleware.RequireRole("admin"), cfg.StudentHandler.GetAll)

		// Endpoint 4: POST /api/v1/students (Admin only)
		students.Post("/", middleware.RequireRole("admin"), cfg.StudentHandler.Create)

		// Endpoint 5: GET /api/v1/students/:id (Admin, Mahasiswa for self)
		students.Get("/:id", cfg.StudentHandler.GetDetail)

		// Endpoint 6: PUT /api/v1/students/:id (Admin only)
		students.Put("/:id", middleware.RequireRole("admin"), cfg.StudentHandler.Update)

		// Endpoint 7: DELETE /api/v1/students/:id (Admin only)
		students.Delete("/:id", middleware.RequireRole("admin"), cfg.StudentHandler.Delete)
	}

	// 3. Course routes (Endpoint 8)
	if cfg.CourseHandler != nil {
		courses := api.Group("/courses")
		courses.Use(middleware.AuthMiddleware(cfg.DB))

		// Endpoint 8: GET /api/v1/courses (All logged-in roles)
		courses.Get("/", cfg.CourseHandler.GetAll)
	}

	// 4. Enrollment routes (Endpoints 9 & 10)
	if cfg.EnrollmentHandler != nil {
		enrollments := api.Group("/enrollments")
		enrollments.Use(middleware.AuthMiddleware(cfg.DB))

		// Endpoint 9: POST /api/v1/enrollments (Mahasiswa only)
		enrollments.Post("/", middleware.RequireRole("mahasiswa"), cfg.EnrollmentHandler.Create)

		// Endpoint 10: DELETE /api/v1/enrollments/:id (Mahasiswa only, own enrollment)
		enrollments.Delete("/:id", middleware.RequireRole("mahasiswa"), cfg.EnrollmentHandler.Delete)
	}
}
