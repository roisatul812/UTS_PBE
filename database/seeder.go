package database

import (
	"database/sql"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type SeedUser struct {
	Email    string
	Password string
	Role     string
}

type SeedStudent struct {
	NIM         string
	Nama        string
	Email       string
	Prodi       string
	Angkatan    int
	IPKTerakhir float64
}

type SeedCourse struct {
	KodeMK   string
	NamaMK   string
	SKS      int
	Semester int
	Kuota    int
}

func Seed(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = 'admin@example.com'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing seed: %w", err)
	}

	if count > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	log.Println("Seeding database...")

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin seed tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Seed Admin
	adminHash, err := bcrypt.GenerateFromPassword([]byte("admin1234"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO users (email, password, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING
	`, "admin@example.com", string(adminHash), "admin")
	if err != nil {
		return fmt.Errorf("failed to seed admin: %w", err)
	}

	// 2. Seed 20 Mahasiswa
	studentsData := []SeedStudent{
		{NIM: "187221000001", Nama: "Rina Putri", Email: "rina.putri@student.siakad.ac.id", Prodi: "Teknik Informatika", Angkatan: 2022, IPKTerakhir: 3.85},
		{NIM: "187221000002", Nama: "Budi Santoso", Email: "budi.santoso@student.siakad.ac.id", Prodi: "Teknik Informatika", Angkatan: 2022, IPKTerakhir: 3.50},
		{NIM: "187221000003", Nama: "Siti Rahma", Email: "siti.rahma@student.siakad.ac.id", Prodi: "Sistem Informasi", Angkatan: 2022, IPKTerakhir: 3.20},
		{NIM: "187221000004", Nama: "Ahmad Fauzi", Email: "ahmad.fauzi@student.siakad.ac.id", Prodi: "Teknik Informatika", Angkatan: 2023, IPKTerakhir: 2.85},
		{NIM: "187221000005", Nama: "Dewi Lestari", Email: "dewi.lestari@student.siakad.ac.id", Prodi: "Sistem Informasi", Angkatan: 2023, IPKTerakhir: 2.70},
		{NIM: "187221000006", Nama: "Eko Prasetyo", Email: "eko.prasetyo@student.siakad.ac.id", Prodi: "Teknik Informatika", Angkatan: 2023, IPKTerakhir: 2.60},
		{NIM: "187221000007", Nama: "Fani Rahmawati", Email: "fani.rahmawati@student.siakad.ac.id", Prodi: "Sistem Informasi", Angkatan: 2024, IPKTerakhir: 2.30},
		{NIM: "187221000008", Nama: "Gilang Permana", Email: "gilang.permana@student.siakad.ac.id", Prodi: "Teknik Informatika", Angkatan: 2024, IPKTerakhir: 2.10},
		{NIM: "187221000009", Nama: "Hani Wijaya", Email: "hani.wijaya@student.siakad.ac.id", Prodi: "Sistem Informasi", Angkatan: 2022, IPKTerakhir: 3.90},
		{NIM: "187221000010", Nama: "Indra Kusuma", Email: "indra.kusuma@student.siakad.ac.id", Prodi: "Teknik Informatika", Angkatan: 2022, IPKTerakhir: 3.40},
		{NIM: "187221000011", Nama: "Joko Susilo", Email: "joko.susilo@student.siakad.ac.id", Prodi: "Teknik Informatika", Angkatan: 2023, IPKTerakhir: 2.95},
		{NIM: "187221000012", Nama: "Kartika Sari", Email: "kartika.sari@student.siakad.ac.id", Prodi: "Sistem Informasi", Angkatan: 2023, IPKTerakhir: 3.10},
		{NIM: "187221000013", Nama: "Lukman Hakim", Email: "lukman.hakim@student.siakad.ac.id", Prodi: "Teknik Informatika", Angkatan: 2024, IPKTerakhir: 2.45},
		{NIM: "187221000014", Nama: "Maya Anggraini", Email: "maya.anggraini@student.siakad.ac.id", Prodi: "Sistem Informasi", Angkatan: 2024, IPKTerakhir: 3.65},
		{NIM: "187221000015", Nama: "Naufal Rizky", Email: "naufal.rizky@student.siakad.ac.id", Prodi: "Teknik Informatika", Angkatan: 2022, IPKTerakhir: 3.75},
		{NIM: "187221000016", Nama: "Olivia Tan", Email: "olivia.tan@student.siakad.ac.id", Prodi: "Sistem Informasi", Angkatan: 2022, IPKTerakhir: 2.80},
		{NIM: "187221000017", Nama: "Panji Nugroho", Email: "panji.nugroho@student.siakad.ac.id", Prodi: "Teknik Informatika", Angkatan: 2023, IPKTerakhir: 2.20},
		{NIM: "187221000018", Nama: "Qori Amelia", Email: "qori.amelia@student.siakad.ac.id", Prodi: "Sistem Informasi", Angkatan: 2023, IPKTerakhir: 3.35},
		{NIM: "187221000019", Nama: "Reza Pratama", Email: "reza.pratama@student.siakad.ac.id", Prodi: "Teknik Informatika", Angkatan: 2024, IPKTerakhir: 3.00},
		{NIM: "187221000020", Nama: "Siska Handayani", Email: "siska.handayani@student.siakad.ac.id", Prodi: "Sistem Informasi", Angkatan: 2024, IPKTerakhir: 2.55},
	}

	for _, s := range studentsData {
		// Initial password is NIM hashed
		pwHash, err := bcrypt.GenerateFromPassword([]byte(s.NIM), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password for %s: %w", s.NIM, err)
		}

		var userID int
		err = tx.QueryRow(`
			INSERT INTO users (email, password, role)
			VALUES ($1, $2, $3)
			RETURNING id
		`, s.Email, string(pwHash), "mahasiswa").Scan(&userID)
		if err != nil {
			return fmt.Errorf("failed to insert user for %s: %w", s.NIM, err)
		}

		_, err = tx.Exec(`
			INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, userID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir)
		if err != nil {
			return fmt.Errorf("failed to insert student for %s: %w", s.NIM, err)
		}
	}

	// 3. Seed 10 Mata Kuliah
	coursesData := []SeedCourse{
		{KodeMK: "IF101", NamaMK: "Pemrograman Dasar", SKS: 3, Semester: 1, Kuota: 30},
		{KodeMK: "IF102", NamaMK: "Struktur Data dan Algoritma", SKS: 3, Semester: 2, Kuota: 30},
		{KodeMK: "IF201", NamaMK: "Basis Data Lanjut", SKS: 3, Semester: 3, Kuota: 25},
		{KodeMK: "IF202", NamaMK: "Pemrograman Web", SKS: 3, Semester: 3, Kuota: 25},
		{KodeMK: "IF301", NamaMK: "Pemrograman Backend Lanjut", SKS: 4, Semester: 5, Kuota: 20},
		{KodeMK: "IF302", NamaMK: "Rekayasa Perangkat Lunak", SKS: 3, Semester: 5, Kuota: 20},
		{KodeMK: "IF303", NamaMK: "Jaringan Komputer", SKS: 3, Semester: 5, Kuota: 25},
		{KodeMK: "IF401", NamaMK: "Kecerdasan Buatan", SKS: 3, Semester: 7, Kuota: 20},
		{KodeMK: "IF402", NamaMK: "Keamanan Informasi", SKS: 3, Semester: 7, Kuota: 2}, // Small quota for testing full quota
		{KodeMK: "IF403", NamaMK: "Cloud Computing", SKS: 2, Semester: 7, Kuota: 1},    // Quota 1 for concurrency test
	}

	for _, c := range coursesData {
		_, err := tx.Exec(`
			INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (kode_mk) DO NOTHING
		`, c.KodeMK, c.NamaMK, c.SKS, c.Semester, c.Kuota)
		if err != nil {
			return fmt.Errorf("failed to insert course %s: %w", c.KodeMK, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit seed: %w", err)
	}

	log.Println("Seeder completed successfully (1 admin, 20 students, 10 courses)")
	return nil
}
