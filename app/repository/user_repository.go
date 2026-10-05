package repository

import (
	"database/sql"
	"errors"

	"uts-pbe-siakad/app/model"
)

type UserRepository interface {
	FindByEmail(email string) (*model.User, error)
	FindByID(id int) (*model.User, error)
	FindStudentProfileByUserID(userID int) (*model.StudentMeProfile, error)
	IsStudentDeleted(userID int) (bool, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	query := `SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = $1`
	row := r.db.QueryRow(query, email)

	var user model.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id int) (*model.User, error) {
	query := `SELECT id, email, password, role, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var user model.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindStudentProfileByUserID(userID int) (*model.StudentMeProfile, error) {
	query := `
		SELECT nim, nama, prodi, angkatan 
		FROM students 
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	row := r.db.QueryRow(query, userID)

	var profile model.StudentMeProfile
	err := row.Scan(&profile.NIM, &profile.Nama, &profile.Prodi, &profile.Angkatan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *userRepository) IsStudentDeleted(userID int) (bool, error) {
	query := `SELECT (deleted_at IS NOT NULL) FROM students WHERE user_id = $1`
	var isDeleted bool
	err := r.db.QueryRow(query, userID).Scan(&isDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return isDeleted, nil
}
