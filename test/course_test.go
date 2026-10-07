package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"uts-pbe-siakad/app/model"
)

func TestCourseGetAllAdminAndMahasiswa(t *testing.T) {
	adminToken := getAdminToken(t)
	studentToken := getStudentToken(t, "rina.putri@student.siakad.ac.id", "187221000001")

	// 1. Admin GET courses
	req := httptest.NewRequest("GET", "/api/v1/courses", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Admin get courses expected 200, got %d", resp.StatusCode)
	}

	var resAdmin struct {
		Success bool           `json:"success"`
		Data    []model.Course `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&resAdmin)
	if len(resAdmin.Data) != 10 {
		t.Errorf("Expected 10 courses, got %d", len(resAdmin.Data))
	}

	// Verify terisi and sisa_kuota calculation
	for _, c := range resAdmin.Data {
		if c.SisaKuota != (c.Kuota - c.Terisi) {
			t.Errorf("Course %s: sisa_kuota (%d) != kuota (%d) - terisi (%d)", c.KodeMK, c.SisaKuota, c.Kuota, c.Terisi)
		}
	}

	// 2. Mahasiswa GET courses
	req = httptest.NewRequest("GET", "/api/v1/courses", nil)
	req.Header.Set("Authorization", "Bearer "+studentToken)
	resp, _ = testApp.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Mahasiswa get courses expected 200, got %d", resp.StatusCode)
	}

	// 3. Unauthenticated GET courses -> 401
	req = httptest.NewRequest("GET", "/api/v1/courses", nil)
	resp, _ = testApp.Test(req, -1)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Unauthenticated get courses expected 401, got %d", resp.StatusCode)
	}
}

func TestCourseFilterSearchAvailable(t *testing.T) {
	adminToken := getAdminToken(t)

	// 1. Search by kode_mk
	req := httptest.NewRequest("GET", "/api/v1/courses?search=IF101", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ := testApp.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Search by kode_mk expected 200, got %d", resp.StatusCode)
	}
	var resSearch struct {
		Data []model.Course `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&resSearch)
	if len(resSearch.Data) != 1 || resSearch.Data[0].KodeMK != "IF101" {
		t.Errorf("Expected to find IF101, got %+v", resSearch.Data)
	}

	// 2. Filter by semester
	req = httptest.NewRequest("GET", "/api/v1/courses?semester=5", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ = testApp.Test(req, -1)
	var resSem struct {
		Data []model.Course `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&resSem)
	if len(resSem.Data) != 3 { // IF301, IF302, IF303 are semester 5
		t.Errorf("Expected 3 courses for semester 5, got %d", len(resSem.Data))
	}

	// 3. Available = true
	req = httptest.NewRequest("GET", "/api/v1/courses?available=true", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, _ = testApp.Test(req, -1)
	var resAvail struct {
		Data []model.Course `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&resAvail)
	for _, c := range resAvail.Data {
		if c.SisaKuota <= 0 {
			t.Errorf("Course %s has sisa_kuota %d <= 0 but returned on available=true", c.KodeMK, c.SisaKuota)
		}
	}
}
