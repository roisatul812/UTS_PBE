package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"uts-pbe-siakad/app/model"
)

func getAdminToken(t *testing.T) string {
	loginPayload := model.LoginRequest{
		Email:    "admin@example.com",
		Password: "admin1234",
	}
	body, _ := json.Marshal(loginPayload)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testApp.Test(req, -1)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to login admin: %v", err)
	}
	var loginResp struct {
		Data model.LoginResponseData `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	return loginResp.Data.AccessToken
}

func getStudentToken(t *testing.T, email, password string) string {
	loginPayload := model.LoginRequest{
		Email:    email,
		Password: password,
	}
	body, _ := json.Marshal(loginPayload)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testApp.Test(req, -1)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to login student: %v", err)
	}
	var loginResp struct {
		Data model.LoginResponseData `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	return loginResp.Data.AccessToken
}

func TestStudentGetAllAdmin(t *testing.T) {
	adminToken := getAdminToken(t)

	req := httptest.NewRequest("GET", "/api/v1/students?page=1&per_page=10", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var res struct {
		Success bool                  `json:"success"`
		Data    []model.Student       `json:"data"`
		Meta    *model.PaginationMeta `json:"meta"`
	}
	json.NewDecoder(resp.Body).Decode(&res)

	if !res.Success {
		t.Errorf("Expected success true")
	}
	if len(res.Data) != 10 {
		t.Errorf("Expected 10 items in page 1, got %d", len(res.Data))
	}
	if res.Meta == nil || res.Meta.Total < 20 {
		t.Errorf("Expected meta total >= 20, got %+v", res.Meta)
	}
}

func TestStudentGetAllForbiddenForMahasiswa(t *testing.T) {
	studentToken := getStudentToken(t, "rina.putri@student.siakad.ac.id", "187221000001")

	req := httptest.NewRequest("GET", "/api/v1/students", nil)
	req.Header.Set("Authorization", "Bearer "+studentToken)
	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for mahasiswa on GET /students, got %d", resp.StatusCode)
	}
}

func TestStudentFilterSearchSort(t *testing.T) {
	adminToken := getAdminToken(t)

	// Search by name
	req := httptest.NewRequest("GET", "/api/v1/students?search=Rina", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ := testApp.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Search failed: got %d", resp.StatusCode)
	}

	var resSearch struct {
		Data []model.Student `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&resSearch)
	if len(resSearch.Data) == 0 || resSearch.Data[0].NIM != "187221000001" {
		t.Errorf("Expected search to find Rina Putri")
	}

	// Filter by prodi and sort by -ipk_terakhir
	req = httptest.NewRequest("GET", "/api/v1/students?prodi=Teknik%20Informatika&sort=-ipk_terakhir", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ = testApp.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Filter sort failed: got %d", resp.StatusCode)
	}

	var resSort struct {
		Data []model.Student `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&resSort)
	if len(resSort.Data) >= 2 {
		if resSort.Data[0].IPKTerakhir < resSort.Data[1].IPKTerakhir {
			t.Errorf("Expected descending sort by IPK, got %f then %f", resSort.Data[0].IPKTerakhir, resSort.Data[1].IPKTerakhir)
		}
	}
}

func TestStudentCreateSuccess(t *testing.T) {
	// Cleanup any leftovers from prior runs
	cleanup := func() {
		testDB.Exec("DELETE FROM students WHERE nim = '187221000099'")
		testDB.Exec("DELETE FROM users WHERE email = 'siswa.baru@student.siakad.ac.id'")
	}
	cleanup()
	defer cleanup()

	adminToken := getAdminToken(t)

	ipk := 3.75
	createPayload := model.CreateStudentRequest{
		NIM:         "187221000099",
		Nama:        "Siswa Baru",
		Email:       "siswa.baru@student.siakad.ac.id",
		Prodi:       "Teknik Informatika",
		Angkatan:    2024,
		IPKTerakhir: &ipk,
	}
	body, _ := json.Marshal(createPayload)

	req := httptest.NewRequest("POST", "/api/v1/students", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", resp.StatusCode)
	}

	// Verify the new student can login with their NIM
	newStudentToken := getStudentToken(t, "siswa.baru@student.siakad.ac.id", "187221000099")
	if newStudentToken == "" {
		t.Errorf("New student should be able to login with password=NIM")
	}
}

func TestStudentCreateValidationFail(t *testing.T) {
	adminToken := getAdminToken(t)

	// Duplicate NIM & Email from seeded student
	ipk := 3.75
	createPayload := model.CreateStudentRequest{
		NIM:         "187221000001",                    // Duplicate
		Nama:        "",                                // Empty
		Email:       "rina.putri@student.siakad.ac.id", // Duplicate
		Prodi:       "Teknik Informatika",
		Angkatan:    2099, // Future year
		IPKTerakhir: &ipk,
	}
	body, _ := json.Marshal(createPayload)

	req := httptest.NewRequest("POST", "/api/v1/students", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, _ := testApp.Test(req, -1)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected 422 for validation fail, got %d", resp.StatusCode)
	}
}

func TestStudentGetDetailAdminAndMahasiswa(t *testing.T) {
	adminToken := getAdminToken(t)
	student1Token := getStudentToken(t, "rina.putri@student.siakad.ac.id", "187221000001")

	// Find student 1 id
	s1, err := studentRepo.FindByNIM("187221000001")
	if err != nil || s1 == nil {
		t.Fatalf("Student 1 not found")
	}
	s2, err := studentRepo.FindByNIM("187221000002")
	if err != nil || s2 == nil {
		t.Fatalf("Student 2 not found")
	}

	// 1. Admin accessing student 1 -> 200
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/students/%d", s1.ID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ := testApp.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Admin get student detail expected 200, got %d", resp.StatusCode)
	}

	var detailRes struct {
		Data model.StudentDetailResponse `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&detailRes)
	if detailRes.Data.BatasSKS != 24 { // IPK 3.85 -> 24 SKS
		t.Errorf("Expected batas_sks 24 for IPK 3.85, got %d", detailRes.Data.BatasSKS)
	}

	// 2. Student 1 accessing self -> 200
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/students/%d", s1.ID), nil)
	req.Header.Set("Authorization", "Bearer "+student1Token)
	resp, _ = testApp.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Student accessing self expected 200, got %d", resp.StatusCode)
	}

	// 3. Student 1 accessing student 2 -> 403 Forbidden!
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/students/%d", s2.ID), nil)
	req.Header.Set("Authorization", "Bearer "+student1Token)
	resp, _ = testApp.Test(req, -1)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Student accessing other student expected 403, got %d", resp.StatusCode)
	}

	// 4. Accessing non-existing student -> 404
	req = httptest.NewRequest("GET", "/api/v1/students/999999", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ = testApp.Test(req, -1)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Accessing non-existing student expected 404, got %d", resp.StatusCode)
	}
}

func TestStudentUpdateAndSoftDelete(t *testing.T) {
	cleanup := func() {
		testDB.Exec("DELETE FROM students WHERE nim = '187221000088'")
		testDB.Exec("DELETE FROM users WHERE email = 'target.delete@student.siakad.ac.id'")
	}
	cleanup()
	defer cleanup()

	adminToken := getAdminToken(t)

	// Create a temporary student to update and delete
	ipk := 2.80
	createPayload := model.CreateStudentRequest{
		NIM:         "187221000088",
		Nama:        "Target Delete",
		Email:       "target.delete@student.siakad.ac.id",
		Prodi:       "Sistem Informasi",
		Angkatan:    2023,
		IPKTerakhir: &ipk,
	}
	body, _ := json.Marshal(createPayload)
	req := httptest.NewRequest("POST", "/api/v1/students", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ := testApp.Test(req, -1)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create student for delete test")
	}

	createdStudent, _ := studentRepo.FindByNIM("187221000088")
	targetID := createdStudent.ID

	// Update student
	newIPK := 3.10
	updatePayload := model.UpdateStudentRequest{
		Nama:        "Target Delete Updated",
		Prodi:       "Teknik Informatika",
		Angkatan:    2023,
		IPKTerakhir: &newIPK,
	}
	upBody, _ := json.Marshal(updatePayload)
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/students/%d", targetID), bytes.NewReader(upBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ = testApp.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Update student expected 200, got %d", resp.StatusCode)
	}

	// Soft delete student -> 204
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/students/%d", targetID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ = testApp.Test(req, -1)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Soft delete expected 204, got %d", resp.StatusCode)
	}

	// Check deleted student cannot login -> 401
	delLoginPayload := model.LoginRequest{
		Email:    "target.delete@student.siakad.ac.id",
		Password: "187221000088",
	}
	delLoginBody, _ := json.Marshal(delLoginPayload)
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(delLoginBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = testApp.Test(req, -1)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Soft-deleted student login expected 401, got %d", resp.StatusCode)
	}

	// Check deleted student not accessible via GET /students/:id -> 404
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/students/%d", targetID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ = testApp.Test(req, -1)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Soft-deleted student detail expected 404, got %d", resp.StatusCode)
	}
}
