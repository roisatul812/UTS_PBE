package service

import (
	"fmt"
	"math"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"uts-pbe-siakad/app/model"
	"uts-pbe-siakad/app/repository"
)

type StudentService interface {
	GetAll(q model.StudentListQuery) ([]model.Student, *model.PaginationMeta, error)
	Create(req model.CreateStudentRequest) (*model.Student, int, string, map[string][]string)
	GetDetail(id int, currentUserID int, currentUserRole string) (*model.StudentDetailResponse, int, string, error)
	Update(id int, req model.UpdateStudentRequest) (*model.Student, int, string, map[string][]string)
	Delete(id int) (int, string, error)
	CalculateBatasSKS(ipk float64) int
}

type studentService struct {
	studentRepo repository.StudentRepository
	userRepo    repository.UserRepository
}

func NewStudentService(studentRepo repository.StudentRepository, userRepo repository.UserRepository) StudentService {
	return &studentService{
		studentRepo: studentRepo,
		userRepo:    userRepo,
	}
}

func (s *studentService) CalculateBatasSKS(ipk float64) int {
	if ipk >= 3.00 {
		return 24
	} else if ipk >= 2.50 {
		return 21
	}
	return 18
}

func (s *studentService) GetAll(q model.StudentListQuery) ([]model.Student, *model.PaginationMeta, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PerPage <= 0 {
		q.PerPage = 10
	} else if q.PerPage > 50 {
		q.PerPage = 50
	}

	students, total, err := s.studentRepo.FindAll(q)
	if err != nil {
		return nil, nil, err
	}

	lastPage := 1
	if total > 0 {
		lastPage = int(math.Ceil(float64(total) / float64(q.PerPage)))
	}

	meta := &model.PaginationMeta{
		CurrentPage: q.Page,
		PerPage:     q.PerPage,
		Total:       total,
		LastPage:    lastPage,
	}

	return students, meta, nil
}

func (s *studentService) Create(req model.CreateStudentRequest) (*model.Student, int, string, map[string][]string) {
	fieldErrors := make(map[string][]string)
	currentYear := time.Now().Year()

	// NIM: required, 12 digits
	matchedNIM, _ := regexp.MatchString(`^[0-9]{12}$`, req.NIM)
	if strings.TrimSpace(req.NIM) == "" {
		fieldErrors["nim"] = append(fieldErrors["nim"], "NIM wajib diisi")
	} else if !matchedNIM {
		fieldErrors["nim"] = append(fieldErrors["nim"], "NIM harus berupa 12 digit angka")
	}

	// Nama: required
	if strings.TrimSpace(req.Nama) == "" {
		fieldErrors["nama"] = append(fieldErrors["nama"], "Nama wajib diisi")
	}

	// Email: required, valid email
	trimmedEmail := strings.TrimSpace(req.Email)
	if trimmedEmail == "" {
		fieldErrors["email"] = append(fieldErrors["email"], "Email wajib diisi")
	} else {
		_, err := mail.ParseAddress(trimmedEmail)
		if err != nil || !strings.Contains(trimmedEmail, "@") || !strings.Contains(trimmedEmail, ".") {
			fieldErrors["email"] = append(fieldErrors["email"], "Format email tidak valid")
		}
	}

	// Prodi: required
	if strings.TrimSpace(req.Prodi) == "" {
		fieldErrors["prodi"] = append(fieldErrors["prodi"], "Program studi wajib diisi")
	}

	// Angkatan: required, 4 digits, <= current year
	if req.Angkatan <= 0 {
		fieldErrors["angkatan"] = append(fieldErrors["angkatan"], "Angkatan wajib diisi")
	} else if req.Angkatan < 1900 || req.Angkatan > 9999 {
		fieldErrors["angkatan"] = append(fieldErrors["angkatan"], "Angkatan harus berupa 4 digit tahun")
	} else if req.Angkatan > currentYear {
		fieldErrors["angkatan"] = append(fieldErrors["angkatan"], fmt.Sprintf("Angkatan tidak boleh melebihi tahun berjalan (%d)", currentYear))
	}

	// IPK Terakhir: optional, 0.00 - 4.00
	if req.IPKTerakhir != nil {
		if *req.IPKTerakhir < 0.00 || *req.IPKTerakhir > 4.00 {
			fieldErrors["ipk_terakhir"] = append(fieldErrors["ipk_terakhir"], "IPK terakhir harus berada di antara 0.00 dan 4.00")
		}
	}

	// Check duplicates before creating
	if len(fieldErrors) == 0 {
		existingStudent, _ := s.studentRepo.FindByNIM(req.NIM)
		if existingStudent != nil {
			fieldErrors["nim"] = append(fieldErrors["nim"], "NIM sudah terdaftar")
		}

		existingUser, _ := s.userRepo.FindByEmail(trimmedEmail)
		if existingUser != nil {
			fieldErrors["email"] = append(fieldErrors["email"], "Email sudah terdaftar")
		}
	}

	if len(fieldErrors) > 0 {
		return nil, 422, "Validasi gagal", fieldErrors
	}

	// Password awal = nim hashed
	hashedPW, err := bcrypt.GenerateFromPassword([]byte(req.NIM), bcrypt.DefaultCost)
	if err != nil {
		return nil, 500, "Gagal mengenkripsi password", nil
	}

	created, err := s.studentRepo.CreateWithUser(req, string(hashedPW))
	if err != nil {
		return nil, 500, "Gagal membuat data mahasiswa: " + err.Error(), nil
	}

	return created, 201, "Mahasiswa berhasil ditambahkan", nil
}

func (s *studentService) GetDetail(id int, currentUserID int, currentUserRole string) (*model.StudentDetailResponse, int, string, error) {
	// If mahasiswa, verify ownership
	if currentUserRole == "mahasiswa" {
		loggedInStudent, err := s.studentRepo.FindByUserID(currentUserID)
		if err != nil {
			return nil, 500, "Terjadi kesalahan internal server", err
		}
		if loggedInStudent == nil || loggedInStudent.ID != id {
			return nil, 403, "Akses ditolak: mahasiswa hanya dapat mengakses data dirinya sendiri", nil
		}
	}

	student, err := s.studentRepo.FindByID(id)
	if err != nil {
		return nil, 500, "Terjadi kesalahan internal server", err
	}
	if student == nil {
		return nil, 404, "Data mahasiswa tidak ditemukan", nil
	}

	courses, totalSKS, err := s.studentRepo.GetEnrolledCourses(id)
	if err != nil {
		return nil, 500, "Gagal mengambil daftar mata kuliah", err
	}

	batasSKS := s.CalculateBatasSKS(student.IPKTerakhir)

	resp := &model.StudentDetailResponse{
		ID:               student.ID,
		NIM:              student.NIM,
		Nama:             student.Nama,
		Email:            student.Email,
		Prodi:            student.Prodi,
		Angkatan:         student.Angkatan,
		IPKTerakhir:      student.IPKTerakhir,
		DaftarMataKuliah: courses,
		TotalSKS:         totalSKS,
		BatasSKS:         batasSKS,
	}

	return resp, 200, "Detail mahasiswa berhasil diambil", nil
}

func (s *studentService) Update(id int, req model.UpdateStudentRequest) (*model.Student, int, string, map[string][]string) {
	// Check existence
	existing, err := s.studentRepo.FindByID(id)
	if err != nil {
		return nil, 500, "Terjadi kesalahan internal server", nil
	}
	if existing == nil {
		return nil, 404, "Data mahasiswa tidak ditemukan", nil
	}

	fieldErrors := make(map[string][]string)
	currentYear := time.Now().Year()

	if strings.TrimSpace(req.Nama) == "" {
		fieldErrors["nama"] = append(fieldErrors["nama"], "Nama wajib diisi")
	}

	if strings.TrimSpace(req.Prodi) == "" {
		fieldErrors["prodi"] = append(fieldErrors["prodi"], "Program studi wajib diisi")
	}

	if req.Angkatan <= 0 {
		fieldErrors["angkatan"] = append(fieldErrors["angkatan"], "Angkatan wajib diisi")
	} else if req.Angkatan < 1900 || req.Angkatan > 9999 {
		fieldErrors["angkatan"] = append(fieldErrors["angkatan"], "Angkatan harus berupa 4 digit tahun")
	} else if req.Angkatan > currentYear {
		fieldErrors["angkatan"] = append(fieldErrors["angkatan"], fmt.Sprintf("Angkatan tidak boleh melebihi tahun berjalan (%d)", currentYear))
	}

	if req.IPKTerakhir != nil {
		if *req.IPKTerakhir < 0.00 || *req.IPKTerakhir > 4.00 {
			fieldErrors["ipk_terakhir"] = append(fieldErrors["ipk_terakhir"], "IPK terakhir harus berada di antara 0.00 dan 4.00")
		}
	}

	if len(fieldErrors) > 0 {
		return nil, 422, "Validasi gagal", fieldErrors
	}

	updated, err := s.studentRepo.Update(id, req)
	if err != nil {
		return nil, 500, "Gagal memperbarui data mahasiswa: " + err.Error(), nil
	}

	return updated, 200, "Data mahasiswa berhasil diperbarui", nil
}

func (s *studentService) Delete(id int) (int, string, error) {
	existing, err := s.studentRepo.FindByID(id)
	if err != nil {
		return 500, "Terjadi kesalahan internal server", err
	}
	if existing == nil {
		return 404, "Data mahasiswa tidak ditemukan", nil
	}

	err = s.studentRepo.SoftDelete(id)
	if err != nil {
		return 500, "Gagal menghapus data mahasiswa", err
	}

	return 204, "Data mahasiswa berhasil dihapus", nil
}
