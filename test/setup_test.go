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
var userRepo repository.UserRepository
var studentRepo repository.StudentRepository
var courseRepo repository.CourseRepository
var enrollmentRepo repository.EnrollmentRepository
var authService service.AuthService
var studentService service.StudentService
var courseService service.CourseService
var enrollmentService service.EnrollmentService

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

	userRepo = repository.NewUserRepository(testDB)
	studentRepo = repository.NewStudentRepository(testDB)
	courseRepo = repository.NewCourseRepository(testDB)
	enrollmentRepo = repository.NewEnrollmentRepository(testDB)

	authService = service.NewAuthService(userRepo)
	studentService = service.NewStudentService(studentRepo, userRepo)
	courseService = service.NewCourseService(courseRepo)
	enrollmentService = service.NewEnrollmentService(enrollmentRepo, studentRepo, studentService)

	authHandler := handler.NewAuthHandler(authService)
	studentHandler := handler.NewStudentHandler(studentService)
	courseHandler := handler.NewCourseHandler(courseService)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentService)

	route.SetupRoutes(&route.RouterConfig{
		App:               testApp,
		DB:                testDB,
		AuthHandler:       authHandler,
		StudentHandler:    studentHandler,
		CourseHandler:     courseHandler,
		EnrollmentHandler: enrollmentHandler,
	})

	code := m.Run()
	testDB.Close()
	os.Exit(code)
}
