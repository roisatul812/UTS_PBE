package middleware

import (
	"database/sql"
	"strings"

	"github.com/gofiber/fiber/v2"
	"uts-pbe-siakad/helper"
)

func AuthMiddleware(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Token autentikasi tidak ditemukan", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Format Authorization header tidak valid", nil)
		}

		tokenString := parts[1]
		claims, err := helper.ValidateToken(tokenString)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Token tidak valid atau telah kedaluwarsa", nil)
		}

		// Verify user still exists in database and check soft delete for student
		var userRole string
		err = db.QueryRow("SELECT role FROM users WHERE id = $1", claims.UserID).Scan(&userRole)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "User tidak ditemukan", nil)
		}

		// If user is mahasiswa, check if soft-deleted
		if userRole == "mahasiswa" {
			var isDeleted bool
			err = db.QueryRow(`
				SELECT (deleted_at IS NOT NULL) 
				FROM students 
				WHERE user_id = $1
			`, claims.UserID).Scan(&isDeleted)
			if err == nil && isDeleted {
				return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Akun mahasiswa ini telah dinonaktifkan", nil)
			}
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}

func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Autentikasi diperlukan", nil)
		}

		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next()
			}
		}

		return helper.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak: role tidak memiliki izin", nil)
	}
}
