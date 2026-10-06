package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"uts-pbe-siakad/app/model"
)

type StudentRepository interface {
	FindAll(q model.StudentListQuery) ([]model.Student, int, error)
	FindByID(id int) (*model.Student, error)
	FindByUserID(userID int) (*model.Student, error)
	FindByNIM(nim string) (*model.Student, error)
	CreateWithUser(req model.CreateStudentRequest, hashedPassword string) (*model.Student, error)
	Update(id int, req model.UpdateStudentRequest) (*model.Student, error)
	SoftDelete(id int) error
	GetEnrolledCourses(studentID int) ([]model.StudentEnrolledCourse, int, error)
}

type studentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) StudentRepository {
	return &studentRepository{db: db}
}

func (r *studentRepository) FindAll(q model.StudentListQuery) ([]model.Student, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	// Exclude soft deleted
	conditions = append(conditions, "s.deleted_at IS NULL")

	if q.Prodi != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(s.prodi) = LOWER($%d)", argIdx))
		args = append(args, q.Prodi)
		argIdx++
	}

	if q.Angkatan > 0 {
		conditions = append(conditions, fmt.Sprintf("s.angkatan = $%d", argIdx))
		args = append(args, q.Angkatan)
		argIdx++
	}

	if q.Search != "" {
		searchPattern := "%" + strings.ToLower(q.Search) + "%"
		conditions = append(conditions, fmt.Sprintf("(LOWER(s.nim) LIKE $%d OR LOWER(s.nama) LIKE $%d)", argIdx, argIdx))
		args = append(args, searchPattern)
		argIdx++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM students s %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Sorting
	orderBy := "s.id ASC"
	switch q.Sort {
	case "nama":
		orderBy = "s.nama ASC"
	case "-nama":
		orderBy = "s.nama DESC"
	case "-ipk_terakhir":
		orderBy = "s.ipk_terakhir DESC"
	case "ipk_terakhir":
		orderBy = "s.ipk_terakhir ASC"
	}

	// Pagination
	limit := q.PerPage
	offset := (q.Page - 1) * q.PerPage

	argsWithPaging := append(args, limit, offset)
	dataQuery := fmt.Sprintf(`
		SELECT s.id, s.user_id, s.nim, s.nama, u.email, s.prodi, s.angkatan, s.ipk_terakhir, s.created_at, s.updated_at
		FROM students s
		JOIN users u ON s.user_id = u.id
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIdx, argIdx+1)

	rows, err := r.db.Query(dataQuery, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Email, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		students = append(students, s)
	}

	if students == nil {
		students = []model.Student{}
	}

	return students, total, nil
}

func (r *studentRepository) FindByID(id int) (*model.Student, error) {
	query := `
		SELECT s.id, s.user_id, s.nim, s.nama, u.email, s.prodi, s.angkatan, s.ipk_terakhir, s.created_at, s.updated_at, s.deleted_at
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = $1 AND s.deleted_at IS NULL
	`
	row := r.db.QueryRow(query, id)

	var s model.Student
	err := row.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Email, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *studentRepository) FindByUserID(userID int) (*model.Student, error) {
	query := `
		SELECT s.id, s.user_id, s.nim, s.nama, u.email, s.prodi, s.angkatan, s.ipk_terakhir, s.created_at, s.updated_at, s.deleted_at
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE s.user_id = $1 AND s.deleted_at IS NULL
	`
	row := r.db.QueryRow(query, userID)

	var s model.Student
	err := row.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Email, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *studentRepository) FindByNIM(nim string) (*model.Student, error) {
	query := `
		SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, created_at, updated_at, deleted_at
		FROM students
		WHERE nim = $1
	`
	row := r.db.QueryRow(query, nim)

	var s model.Student
	err := row.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *studentRepository) CreateWithUser(req model.CreateStudentRequest, hashedPassword string) (*model.Student, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Insert user
	var userID int
	err = tx.QueryRow(`
		INSERT INTO users (email, password, role)
		VALUES ($1, $2, $3)
		RETURNING id
	`, req.Email, hashedPassword, "mahasiswa").Scan(&userID)
	if err != nil {
		return nil, err
	}

	ipk := 0.00
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
	}

	// 2. Insert student
	var s model.Student
	err = tx.QueryRow(`
		INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, created_at, updated_at
	`, userID, req.NIM, req.Nama, req.Prodi, req.Angkatan, ipk).Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	s.Email = req.Email

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *studentRepository) Update(id int, req model.UpdateStudentRequest) (*model.Student, error) {
	// First check if student exists and not deleted
	student, err := r.FindByID(id)
	if err != nil || student == nil {
		return nil, err
	}

	ipk := student.IPKTerakhir
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
	}

	var updated model.Student
	query := `
		UPDATE students
		SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5 AND deleted_at IS NULL
		RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, created_at, updated_at
	`
	err = r.db.QueryRow(query, req.Nama, req.Prodi, req.Angkatan, ipk, id).Scan(
		&updated.ID, &updated.UserID, &updated.NIM, &updated.Nama, &updated.Prodi, &updated.Angkatan, &updated.IPKTerakhir, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	updated.Email = student.Email

	return &updated, nil
}

func (r *studentRepository) SoftDelete(id int) error {
	result, err := r.db.Exec(`
		UPDATE students 
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *studentRepository) GetEnrolledCourses(studentID int) ([]model.StudentEnrolledCourse, int, error) {
	query := `
		SELECT e.id, e.course_id, c.kode_mk, c.nama_mk, c.sks, c.semester, e.tahun_akademik
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.student_id = $1
		ORDER BY e.created_at ASC
	`
	rows, err := r.db.Query(query, studentID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.StudentEnrolledCourse
	totalSKS := 0
	for rows.Next() {
		var item model.StudentEnrolledCourse
		if err := rows.Scan(&item.EnrollmentID, &item.CourseID, &item.KodeMK, &item.NamaMK, &item.SKS, &item.Semester, &item.TahunAkademik); err != nil {
			return nil, 0, err
		}
		totalSKS += item.SKS
		list = append(list, item)
	}

	if list == nil {
		list = []model.StudentEnrolledCourse{}
	}

	return list, totalSKS, nil
}
