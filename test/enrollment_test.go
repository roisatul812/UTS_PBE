package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"uts-pbe-siakad/app/model"
)

func TestEnrollmentSuccess(t *testing.T) {
	cleanup := func() {
		testDB.Exec("DELETE FROM enrollments")
	}
	cleanup()
	defer cleanup()

	studentToken := getStudentToken(t, "rina.putri@student.siakad.ac.id", "187221000001")

	// Get a course
	c, err := courseRepo.FindByID(1)
	if err != nil || c == nil {
		t.Fatalf("Course 1 not found")
	}

	payload := model.CreateEnrollmentRequest{
		CourseID:      c.ID,
		TahunAkademik: "2026/2027-Ganjil",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/enrollments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+studentToken)

	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", resp.StatusCode)
	}

	var res struct {
		Success bool                         `json:"success"`
		Data    model.EnrollmentResponseData `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	if res.Data.CourseID != c.ID || res.Data.TahunAkademik != "2026/2027-Ganjil" {
		t.Errorf("Unexpected enrollment response: %+v", res.Data)
	}
}

func TestEnrollmentDuplicateCourse(t *testing.T) {
	cleanup := func() {
		testDB.Exec("DELETE FROM enrollments")
	}
	cleanup()
	defer cleanup()

	studentToken := getStudentToken(t, "rina.putri@student.siakad.ac.id", "187221000001")

	payload := model.CreateEnrollmentRequest{
		CourseID:      1,
		TahunAkademik: "2026/2027-Ganjil",
	}
	body, _ := json.Marshal(payload)

	// First enrollment -> 201
	req1 := httptest.NewRequest("POST", "/api/v1/enrollments", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer "+studentToken)
	resp1, _ := testApp.Test(req1, -1)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("First enrollment failed: %d", resp1.StatusCode)
	}

	// Second enrollment same course and semester -> 409 Conflict
	req2 := httptest.NewRequest("POST", "/api/v1/enrollments", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+studentToken)
	resp2, _ := testApp.Test(req2, -1)
	if resp2.StatusCode != http.StatusConflict {
		t.Errorf("Expected 409 Conflict on duplicate course, got %d", resp2.StatusCode)
	}
}

func TestEnrollmentQuotaFull(t *testing.T) {
	cleanup := func() {
		testDB.Exec("DELETE FROM enrollments")
	}
	cleanup()
	defer cleanup()

	// Find course with quota = 1 (e.g. IF403 Cloud Computing)
	var smallCourseID int
	var kuota int
	err := testDB.QueryRow("SELECT id, kuota FROM courses WHERE kuota = 1 LIMIT 1").Scan(&smallCourseID, &kuota)
	if err != nil {
		t.Fatalf("No course with kuota 1 found: %v", err)
	}

	student1Token := getStudentToken(t, "rina.putri@student.siakad.ac.id", "187221000001")
	student2Token := getStudentToken(t, "budi.santoso@student.siakad.ac.id", "187221000002")

	payload := model.CreateEnrollmentRequest{
		CourseID:      smallCourseID,
		TahunAkademik: "2026/2027-Ganjil",
	}
	body, _ := json.Marshal(payload)

	// Student 1 enrolls -> 201 (fills quota 1 of 1)
	req1 := httptest.NewRequest("POST", "/api/v1/enrollments", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer "+student1Token)
	resp1, _ := testApp.Test(req1, -1)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("Student 1 enrollment failed: %d", resp1.StatusCode)
	}

	// Student 2 tries to enroll same course -> 422 Kuota Penuh
	req2 := httptest.NewRequest("POST", "/api/v1/enrollments", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+student2Token)
	resp2, _ := testApp.Test(req2, -1)
	if resp2.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected 422 for full quota, got %d", resp2.StatusCode)
	}

	var res2 model.ApiResponse
	json.NewDecoder(resp2.Body).Decode(&res2)
	if !strings.Contains(strings.ToLower(res2.Message), "kuota") {
		t.Errorf("Expected message to mention 'kuota', got '%s'", res2.Message)
	}
}

func TestEnrollmentSKSExceeded(t *testing.T) {
	cleanup := func() {
		testDB.Exec("DELETE FROM enrollments")
	}
	cleanup()
	defer cleanup()

	// Student 8: Gilang Permana (NIM 187221000008, IPK 2.10 -> batas_sks = 18 SKS)
	token := getStudentToken(t, "gilang.permana@student.siakad.ac.id", "187221000008")

	// Get courses with known SKS
	rows, err := testDB.Query("SELECT id, sks FROM courses ORDER BY id ASC")
	if err != nil {
		t.Fatalf("Failed to query courses: %v", err)
	}
	defer rows.Close()

	type courseInfo struct {
		id  int
		sks int
	}
	var courseList []courseInfo
	for rows.Next() {
		var ci courseInfo
		rows.Scan(&ci.id, &ci.sks)
		courseList = append(courseList, ci)
	}

	// Enroll up to 18 SKS
	accumulatedSKS := 0
	tahunAkademik := "2026/2027-Ganjil"

	var exceededCourse courseInfo
	for _, c := range courseList {
		if accumulatedSKS+c.sks <= 18 {
			payload := model.CreateEnrollmentRequest{
				CourseID:      c.id,
				TahunAkademik: tahunAkademik,
			}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/api/v1/enrollments", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			resp, _ := testApp.Test(req, -1)
			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("Failed to enroll course %d: %d", c.id, resp.StatusCode)
			}
			accumulatedSKS += c.sks
		} else {
			exceededCourse = c
			break
		}
	}

	// Now try to enroll exceededCourse (which will exceed 18 SKS)
	payload := model.CreateEnrollmentRequest{
		CourseID:      exceededCourse.id,
		TahunAkademik: tahunAkademik,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/enrollments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := testApp.Test(req, -1)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected 422 for SKS exceeded, got %d", resp.StatusCode)
	}

	var res model.ApiResponse
	json.NewDecoder(resp.Body).Decode(&res)
	// Must mention sisa SKS
	if !strings.Contains(strings.ToLower(res.Message), "sisa sks") {
		t.Errorf("Expected error message to mention 'sisa SKS', got '%s'", res.Message)
	}
}

func TestEnrollmentForbiddenForAdmin(t *testing.T) {
	adminToken := getAdminToken(t)

	payload := model.CreateEnrollmentRequest{
		CourseID:      1,
		TahunAkademik: "2026/2027-Ganjil",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/v1/enrollments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, _ := testApp.Test(req, -1)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for admin enrolling course, got %d", resp.StatusCode)
	}
}

func TestUnenrollSuccessAndQuotaReturned(t *testing.T) {
	cleanup := func() {
		testDB.Exec("DELETE FROM enrollments")
	}
	cleanup()
	defer cleanup()

	studentToken := getStudentToken(t, "rina.putri@student.siakad.ac.id", "187221000001")

	// Enroll course 1
	payload := model.CreateEnrollmentRequest{
		CourseID:      1,
		TahunAkademik: "2026/2027-Ganjil",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/enrollments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+studentToken)
	resp, _ := testApp.Test(req, -1)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Enroll failed: %d", resp.StatusCode)
	}

	var enrollRes struct {
		Data model.EnrollmentResponseData `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&enrollRes)
	enrollmentID := enrollRes.Data.ID

	// Check sisa_kuota decreased
	cBefore, _ := courseRepo.FindByID(1)

	// Unenroll (DELETE /api/v1/enrollments/:id)
	delReq := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/enrollments/%d", enrollmentID), nil)
	delReq.Header.Set("Authorization", "Bearer "+studentToken)
	delResp, _ := testApp.Test(delReq, -1)
	if delResp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected 204 No Content on unenroll, got %d", delResp.StatusCode)
	}

	// Check sisa_kuota increased back
	cAfter, _ := courseRepo.FindByID(1)
	if cAfter.SisaKuota != cBefore.SisaKuota+1 {
		t.Errorf("Expected sisa_kuota to increase from %d to %d, got %d", cBefore.SisaKuota, cBefore.SisaKuota+1, cAfter.SisaKuota)
	}
}

func TestUnenrollForbiddenOtherStudentAndNotFound(t *testing.T) {
	cleanup := func() {
		testDB.Exec("DELETE FROM enrollments")
	}
	cleanup()
	defer cleanup()

	student1Token := getStudentToken(t, "rina.putri@student.siakad.ac.id", "187221000001")
	student2Token := getStudentToken(t, "budi.santoso@student.siakad.ac.id", "187221000002")

	// Student 1 enrolls
	payload := model.CreateEnrollmentRequest{
		CourseID:      1,
		TahunAkademik: "2026/2027-Ganjil",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/enrollments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+student1Token)
	resp, _ := testApp.Test(req, -1)
	var enrollRes struct {
		Data model.EnrollmentResponseData `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&enrollRes)
	enrollmentID := enrollRes.Data.ID

	// Student 2 tries to delete Student 1's enrollment -> 403 Forbidden!
	delReq := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/enrollments/%d", enrollmentID), nil)
	delReq.Header.Set("Authorization", "Bearer "+student2Token)
	delResp, _ := testApp.Test(delReq, -1)
	if delResp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden when deleting other student enrollment, got %d", delResp.StatusCode)
	}

	// Delete non-existent enrollment -> 404 Not Found
	delReq404 := httptest.NewRequest("DELETE", "/api/v1/enrollments/999999", nil)
	delReq404.Header.Set("Authorization", "Bearer "+student1Token)
	delResp404, _ := testApp.Test(delReq404, -1)
	if delResp404.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found for non-existing enrollment, got %d", delResp404.StatusCode)
	}
}
