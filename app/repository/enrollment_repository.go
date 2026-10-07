package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"uts-pbe-siakad/app/model"
)

type EnrollmentRepository interface {
	FindByID(id int) (*model.Enrollment, error)
	Delete(id int) error
	CreateWithLocking(studentID int, batasSKS int, req model.CreateEnrollmentRequest) (*model.EnrollmentResponseData, int, string, error)
}

type enrollmentRepository struct {
	db *sql.DB
}

func NewEnrollmentRepository(db *sql.DB) EnrollmentRepository {
	return &enrollmentRepository{db: db}
}

func (r *enrollmentRepository) FindByID(id int) (*model.Enrollment, error) {
	query := `SELECT id, student_id, course_id, tahun_akademik, created_at FROM enrollments WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var e model.Enrollment
	err := row.Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *enrollmentRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM enrollments WHERE id = $1", id)
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

func (r *enrollmentRepository) CreateWithLocking(studentID int, batasSKS int, req model.CreateEnrollmentRequest) (*model.EnrollmentResponseData, int, string, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, 500, "Gagal memulai transaksi", err
	}
	defer tx.Rollback()

	// 1. Lock course row to prevent race conditions on quota
	var course model.Course
	err = tx.QueryRow(`
		SELECT id, kode_mk, nama_mk, sks, kuota 
		FROM courses 
		WHERE id = $1 
		FOR UPDATE
	`, req.CourseID).Scan(&course.ID, &course.KodeMK, &course.NamaMK, &course.SKS, &course.Kuota)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 404, "Mata kuliah tidak ditemukan", nil
		}
		return nil, 500, "Gagal memverifikasi mata kuliah", err
	}

	// 2. Check if student already enrolled in this course in the same tahun_akademik
	var duplicateCount int
	err = tx.QueryRow(`
		SELECT COUNT(*) 
		FROM enrollments 
		WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3
	`, studentID, req.CourseID, req.TahunAkademik).Scan(&duplicateCount)
	if err != nil {
		return nil, 500, "Gagal mengecek duplikasi mata kuliah", err
	}
	if duplicateCount > 0 {
		return nil, 409, "Mata kuliah sudah pernah diambil pada tahun akademik yang sama", nil
	}

	// 3. Check quota with current count
	var currentCourseEnrolled int
	err = tx.QueryRow(`
		SELECT COUNT(*) 
		FROM enrollments 
		WHERE course_id = $1
	`, req.CourseID).Scan(&currentCourseEnrolled)
	if err != nil {
		return nil, 500, "Gagal mengecek kuota mata kuliah", err
	}
	if currentCourseEnrolled >= course.Kuota {
		return nil, 422, "Kuota mata kuliah sudah penuh", nil
	}

	// 4. Calculate total SKS taken by student in this tahun_akademik
	var currentTotalSKS int
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(c.sks), 0) 
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.student_id = $1 AND e.tahun_akademik = $2
	`, studentID, req.TahunAkademik).Scan(&currentTotalSKS)
	if err != nil {
		return nil, 500, "Gagal menghitung total SKS", err
	}

	// 5. Validate SKS limit
	sisaSKS := batasSKS - currentTotalSKS
	if sisaSKS < 0 {
		sisaSKS = 0
	}

	if currentTotalSKS+course.SKS > batasSKS {
		msg := fmt.Sprintf("Total SKS melebihi batas. Sisa SKS yang masih tersedia: %d SKS", sisaSKS)
		return nil, 422, msg, nil
	}

	// 6. Insert enrollment
	var enrolled model.EnrollmentResponseData
	err = tx.QueryRow(`
		INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		VALUES ($1, $2, $3)
		RETURNING id, student_id, course_id, tahun_akademik, created_at
	`, studentID, req.CourseID, req.TahunAkademik).Scan(
		&enrolled.ID, &enrolled.StudentID, &enrolled.CourseID, &enrolled.TahunAkademik, &enrolled.CreatedAt,
	)
	if err != nil {
		return nil, 500, "Gagal menyimpan KRS", err
	}

	enrolled.KodeMK = course.KodeMK
	enrolled.NamaMK = course.NamaMK
	enrolled.SKS = course.SKS

	// 7. Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, 500, "Gagal menyelesaikan transaksi pendaftaran KRS", err
	}

	return &enrolled, 201, "Mata kuliah berhasil ditambahkan ke KRS", nil
}
