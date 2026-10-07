package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"uts-pbe-siakad/app/model"
)

type CourseRepository interface {
	FindAll(q model.CourseListQuery) ([]model.Course, error)
	FindByID(id int) (*model.Course, error)
}

type courseRepository struct {
	db *sql.DB
}

func NewCourseRepository(db *sql.DB) CourseRepository {
	return &courseRepository{db: db}
}

func (r *courseRepository) FindAll(q model.CourseListQuery) ([]model.Course, error) {
	var whereConditions []string
	var args []interface{}
	argIdx := 1

	if q.Semester > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("c.semester = $%d", argIdx))
		args = append(args, q.Semester)
		argIdx++
	}

	if q.Search != "" {
		searchPattern := "%" + strings.ToLower(q.Search) + "%"
		whereConditions = append(whereConditions, fmt.Sprintf("(LOWER(c.kode_mk) LIKE $%d OR LOWER(c.nama_mk) LIKE $%d)", argIdx, argIdx))
		args = append(args, searchPattern)
		argIdx++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	havingClause := ""
	if q.Available {
		havingClause = "HAVING (c.kuota - COUNT(e.id)) > 0"
	}

	query := fmt.Sprintf(`
		SELECT 
			c.id, 
			c.kode_mk, 
			c.nama_mk, 
			c.sks, 
			c.semester, 
			c.kuota,
			COUNT(e.id)::int AS terisi,
			(c.kuota - COUNT(e.id))::int AS sisa_kuota,
			c.created_at,
			c.updated_at
		FROM courses c
		LEFT JOIN enrollments e ON c.id = e.course_id
		%s
		GROUP BY c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, c.created_at, c.updated_at
		%s
		ORDER BY c.semester ASC, c.kode_mk ASC
	`, whereClause, havingClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []model.Course
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(
			&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota,
			&c.Terisi, &c.SisaKuota, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}

	if courses == nil {
		courses = []model.Course{}
	}

	return courses, nil
}

func (r *courseRepository) FindByID(id int) (*model.Course, error) {
	query := `
		SELECT 
			c.id, 
			c.kode_mk, 
			c.nama_mk, 
			c.sks, 
			c.semester, 
			c.kuota,
			COUNT(e.id)::int AS terisi,
			(c.kuota - COUNT(e.id))::int AS sisa_kuota,
			c.created_at,
			c.updated_at
		FROM courses c
		LEFT JOIN enrollments e ON c.id = e.course_id
		WHERE c.id = $1
		GROUP BY c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, c.created_at, c.updated_at
	`
	row := r.db.QueryRow(query, id)

	var c model.Course
	err := row.Scan(
		&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota,
		&c.Terisi, &c.SisaKuota, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}
