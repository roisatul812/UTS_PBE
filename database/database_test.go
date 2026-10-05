package database

import (
	"testing"

	"uts-pbe-siakad/config"
)

func TestDatabaseMigrationAndSeeder(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := ConnectDB(cfg)
	if err != nil {
		t.Fatalf("Failed to connect database: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db, "../migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	if err := Seed(db); err != nil {
		t.Fatalf("Failed to run seed: %v", err)
	}

	// Verify Admin
	var adminCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount)
	if err != nil {
		t.Fatalf("Failed to query admin: %v", err)
	}
	if adminCount < 1 {
		t.Errorf("Expected at least 1 admin, got %d", adminCount)
	}

	// Verify Students
	var studentCount int
	err = db.QueryRow("SELECT COUNT(*) FROM students").Scan(&studentCount)
	if err != nil {
		t.Fatalf("Failed to query students: %v", err)
	}
	if studentCount != 20 {
		t.Errorf("Expected exactly 20 students, got %d", studentCount)
	}

	// Verify Courses
	var courseCount int
	err = db.QueryRow("SELECT COUNT(*) FROM courses").Scan(&courseCount)
	if err != nil {
		t.Fatalf("Failed to query courses: %v", err)
	}
	if courseCount != 10 {
		t.Errorf("Expected exactly 10 courses, got %d", courseCount)
	}
}
