package model

import "time"

type Course struct {
	ID        int       `json:"id"`
	KodeMK    string    `json:"kode_mk"`
	NamaMK    string    `json:"nama_mk"`
	SKS       int       `json:"sks"`
	Semester  int       `json:"semester"`
	Kuota     int       `json:"kuota"`
	Terisi    int       `json:"terisi"`
	SisaKuota int       `json:"sisa_kuota"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type CourseListQuery struct {
	Semester  int
	Search    string
	Available bool
}
