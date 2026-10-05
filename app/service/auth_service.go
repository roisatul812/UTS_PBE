package service

import (
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"uts-pbe-siakad/app/model"
	"uts-pbe-siakad/app/repository"
	"uts-pbe-siakad/config"
	"uts-pbe-siakad/helper"
	"uts-pbe-siakad/middleware"
)

type AuthService interface {
	Login(req model.LoginRequest, clientIP string) (*model.LoginResponseData, int, string, map[string][]string)
	GetMe(userID int) (*model.AuthMeResponseData, int, string, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Login(req model.LoginRequest, clientIP string) (*model.LoginResponseData, int, string, map[string][]string) {
	// 1. Validation
	fieldErrors := make(map[string][]string)

	trimmedEmail := strings.TrimSpace(req.Email)
	if trimmedEmail == "" {
		fieldErrors["email"] = append(fieldErrors["email"], "Email wajib diisi")
	} else {
		_, err := mail.ParseAddress(trimmedEmail)
		if err != nil || !strings.Contains(trimmedEmail, "@") || !strings.Contains(trimmedEmail, ".") {
			fieldErrors["email"] = append(fieldErrors["email"], "Format email tidak valid")
		}
	}

	if req.Password == "" {
		fieldErrors["password"] = append(fieldErrors["password"], "Password wajib diisi")
	} else if len(req.Password) < 8 {
		fieldErrors["password"] = append(fieldErrors["password"], "Password minimal 8 karakter")
	}

	if len(fieldErrors) > 0 {
		return nil, 422, "Validasi gagal", fieldErrors
	}

	// 2. Check Rate Limit (failed login > 5 per minute)
	rateLimitKey := clientIP + ":" + trimmedEmail
	if middleware.LoginLimiter.IsBlocked(rateLimitKey) || middleware.LoginLimiter.IsBlocked(clientIP) {
		return nil, 429, "Terlalu banyak percobaan login gagal, silakan coba lagi dalam 1 menit", nil
	}

	// 3. Find User
	user, err := s.userRepo.FindByEmail(trimmedEmail)
	if err != nil || user == nil {
		middleware.LoginLimiter.RecordFailure(rateLimitKey)
		middleware.LoginLimiter.RecordFailure(clientIP)
		return nil, 401, "Email atau password salah", nil
	}

	// 4. Verify Password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		middleware.LoginLimiter.RecordFailure(rateLimitKey)
		middleware.LoginLimiter.RecordFailure(clientIP)
		return nil, 401, "Email atau password salah", nil
	}

	// 5. Check if student is soft-deleted
	if user.Role == "mahasiswa" {
		isDeleted, err := s.userRepo.IsStudentDeleted(user.ID)
		if err != nil || isDeleted {
			return nil, 401, "Akun mahasiswa ini telah dinonaktifkan", nil
		}
	}

	// Success: Reset rate limiter
	middleware.LoginLimiter.Reset(rateLimitKey)
	middleware.LoginLimiter.Reset(clientIP)

	// 6. Generate Token
	token, err := helper.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, 500, "Gagal membuat access token", nil
	}

	cfg := config.AppConfig
	resp := &model.LoginResponseData{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   cfg.JWTExpiresIn,
		User: model.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		},
	}

	return resp, 200, "Login berhasil", nil
}

func (s *authService) GetMe(userID int) (*model.AuthMeResponseData, int, string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, 500, "Terjadi kesalahan internal server", err
	}
	if user == nil {
		return nil, 401, "User tidak ditemukan", nil
	}

	data := &model.AuthMeResponseData{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}

	if user.Role == "mahasiswa" {
		profile, err := s.userRepo.FindStudentProfileByUserID(userID)
		if err != nil {
			return nil, 500, "Terjadi kesalahan internal server", err
		}
		data.Student = profile
	}

	return data, 200, "Profil pengguna berhasil diambil", nil
}
