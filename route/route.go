package route

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"uts-pbe-siakad/app/handler"
	"uts-pbe-siakad/middleware"
)

type RouterConfig struct {
	App            *fiber.App
	DB             *sql.DB
	AuthHandler    *handler.AuthHandler
	StudentHandler *handler.StudentHandler
}

func SetupRoutes(cfg *RouterConfig) {
	api := cfg.App.Group("/api/v1")

	// 1. Auth routes
	auth := api.Group("/auth")
	auth.Post("/login", cfg.AuthHandler.Login)
	auth.Get("/me", middleware.AuthMiddleware(cfg.DB), cfg.AuthHandler.GetMe)

	// 2. Student routes
	if cfg.StudentHandler != nil {
		students := api.Group("/students")
		students.Use(middleware.AuthMiddleware(cfg.DB))

		// GET /api/v1/students (Admin only)
		students.Get("/", middleware.RequireRole("admin"), cfg.StudentHandler.GetAll)

		// POST /api/v1/students (Admin only)
		students.Post("/", middleware.RequireRole("admin"), cfg.StudentHandler.Create)

		// GET /api/v1/students/:id (Admin, Mahasiswa for self)
		students.Get("/:id", cfg.StudentHandler.GetDetail)

		// PUT /api/v1/students/:id (Admin only)
		students.Put("/:id", middleware.RequireRole("admin"), cfg.StudentHandler.Update)

		// DELETE /api/v1/students/:id (Admin only)
		students.Delete("/:id", middleware.RequireRole("admin"), cfg.StudentHandler.Delete)
	}
}
