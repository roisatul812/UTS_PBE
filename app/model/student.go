package model

import "time"

type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Email       string     `json:"email,omitempty"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateStudentRequest struct {
	NIM         string   `json:"nim"`
	Nama        string   `json:"nama"`
	Email       string   `json:"email"`
	Prodi       string   `json:"prodi"`
	Angkatan    int      `json:"angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty"`
}

type UpdateStudentRequest struct {
	Nama        string   `json:"nama"`
	Prodi       string   `json:"prodi"`
	Angkatan    int      `json:"angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir,omitempty"`
}

type StudentEnrolledCourse struct {
	EnrollmentID  int    `json:"enrollment_id"`
	CourseID      int    `json:"course_id"`
	KodeMK        string `json:"kode_mk"`
	NamaMK        string `json:"nama_mk"`
	SKS           int    `json:"sks"`
	Semester      int    `json:"semester"`
	TahunAkademik string `json:"tahun_akademik"`
}

type StudentDetailResponse struct {
	ID               int                     `json:"id"`
	NIM              string                  `json:"nim"`
	Nama             string                  `json:"nama"`
	Email            string                  `json:"email,omitempty"`
	Prodi            string                  `json:"prodi"`
	Angkatan         int                     `json:"angkatan"`
	IPKTerakhir      float64                 `json:"ipk_terakhir"`
	DaftarMataKuliah []StudentEnrolledCourse `json:"daftar_mata_kuliah"`
	TotalSKS         int                     `json:"total_sks"`
	BatasSKS         int                     `json:"batas_sks"`
}

type StudentListQuery struct {
	Page     int
	PerPage  int
	Prodi    string
	Angkatan int
	Search   string
	Sort     string
}
