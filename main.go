package main

import (
	"flag"
	"log"

	"github.com/gofiber/fiber/v2"
	"uts-pbe-siakad/config"
	"uts-pbe-siakad/database"
)

func main() {
	seedFlag := flag.Bool("seed", false, "Run database seed")
	flag.Parse()

	cfg := config.LoadConfig()

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db, "migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if *seedFlag {
		if err := database.Seed(db); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
	} else {
		// Auto seed if users table is empty
		if err := database.Seed(db); err != nil {
			log.Printf("Seed check/run error: %v", err)
		}
	}

	app := fiber.New()

	log.Printf("Server starting on port %s...", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
