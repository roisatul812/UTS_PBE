package test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"uts-pbe-siakad/app/handler"
	"uts-pbe-siakad/app/repository"
	"uts-pbe-siakad/app/service"
	"uts-pbe-siakad/config"
	"uts-pbe-siakad/database"
	"uts-pbe-siakad/route"
)

var testApp *fiber.App
var testDB *sql.DB

func TestMain(m *testing.M) {
	cfg := config.LoadConfig()
	var err error
	testDB, err = database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Test DB connection failed: %v", err)
	}

	if err := database.RunMigrations(testDB, "../migrations"); err != nil {
		log.Fatalf("Test DB migration failed: %v", err)
	}

	if err := database.Seed(testDB); err != nil {
		log.Fatalf("Test DB seed failed: %v", err)
	}

	testApp = fiber.New()

	userRepo := repository.NewUserRepository(testDB)
	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)

	route.SetupRoutes(&route.RouterConfig{
		App:         testApp,
		DB:          testDB,
		AuthHandler: authHandler,
	})

	code := m.Run()
	testDB.Close()
	os.Exit(code)
}
