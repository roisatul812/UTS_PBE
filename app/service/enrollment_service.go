package service

import (
	"strings"

	"uts-pbe-siakad/app/model"
	"uts-pbe-siakad/app/repository"
)

type EnrollmentService interface {
	Enroll(userID int, req model.CreateEnrollmentRequest) (*model.EnrollmentResponseData, int, string, map[string][]string)
	Unenroll(userID int, enrollmentID int) (int, string, error)
}

type enrollmentService struct {
	enrollmentRepo repository.EnrollmentRepository
	studentRepo    repository.StudentRepository
	studentService StudentService
}

func NewEnrollmentService(
	enrollmentRepo repository.EnrollmentRepository,
	studentRepo repository.StudentRepository,
	studentService StudentService,
) EnrollmentService {
	return &enrollmentService{
		enrollmentRepo: enrollmentRepo,
		studentRepo:    studentRepo,
		studentService: studentService,
	}
}

func (s *enrollmentService) Enroll(userID int, req model.CreateEnrollmentRequest) (*model.EnrollmentResponseData, int, string, map[string][]string) {
	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil || student == nil {
		return nil, 403, "Akses ditolak: profil mahasiswa tidak ditemukan", nil
	}

	// Validate input
	fieldErrors := make(map[string][]string)
	if req.CourseID <= 0 {
		fieldErrors["course_id"] = append(fieldErrors["course_id"], "ID mata kuliah wajib diisi")
	}

	if strings.TrimSpace(req.TahunAkademik) == "" {
		fieldErrors["tahun_akademik"] = append(fieldErrors["tahun_akademik"], "Tahun akademik wajib diisi")
	}

	if len(fieldErrors) > 0 {
		return nil, 422, "Validasi gagal", fieldErrors
	}

	batasSKS := s.studentService.CalculateBatasSKS(student.IPKTerakhir)

	enrolled, statusCode, message, err := s.enrollmentRepo.CreateWithLocking(student.ID, batasSKS, req)
	if statusCode != 201 {
		return nil, statusCode, message, nil
	}

	return enrolled, 201, message, nil
}

func (s *enrollmentService) Unenroll(userID int, enrollmentID int) (int, string, error) {
	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil || student == nil {
		return 403, "Akses ditolak: profil mahasiswa tidak ditemukan", nil
	}

	enrollment, err := s.enrollmentRepo.FindByID(enrollmentID)
	if err != nil {
		return 500, "Terjadi kesalahan internal server", err
	}
	if enrollment == nil {
		return 404, "Data enrollment tidak ditemukan", nil
	}

	if enrollment.StudentID != student.ID {
		return 403, "Akses ditolak: mahasiswa hanya dapat membatalkan KRS miliknya sendiri", nil
	}

	if err := s.enrollmentRepo.Delete(enrollmentID); err != nil {
		return 500, "Gagal membatalkan mata kuliah dari KRS", err
	}

	return 204, "Mata kuliah berhasil dibatalkan dari KRS", nil
}
