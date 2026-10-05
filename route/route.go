package route

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"uts-pbe-siakad/app/handler"
	"uts-pbe-siakad/middleware"
)

type RouterConfig struct {
	App         *fiber.App
	DB          *sql.DB
	AuthHandler *handler.AuthHandler
}

func SetupRoutes(cfg *RouterConfig) {
	api := cfg.App.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/login", cfg.AuthHandler.Login)
	auth.Get("/me", middleware.AuthMiddleware(cfg.DB), cfg.AuthHandler.GetMe)
}
