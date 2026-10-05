package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"uts-pbe-siakad/app/model"
	"uts-pbe-siakad/middleware"
)

func TestAuthLoginAdminSuccess(t *testing.T) {
	loginPayload := model.LoginRequest{
		Email:    "admin@example.com",
		Password: "admin1234",
	}
	body, _ := json.Marshal(loginPayload)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var res model.ApiResponse
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Success {
		t.Errorf("Expected success true, got false: %s", res.Message)
	}
}

func TestAuthLoginMahasiswaSuccess(t *testing.T) {
	// First student in seeder: 187221000001 / rina.putri@student.siakad.ac.id, initial password is NIM
	loginPayload := model.LoginRequest{
		Email:    "rina.putri@student.siakad.ac.id",
		Password: "187221000001",
	}
	body, _ := json.Marshal(loginPayload)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestAuthLoginValidationFail(t *testing.T) {
	// Invalid email and password < 8 chars
	loginPayload := model.LoginRequest{
		Email:    "not-an-email",
		Password: "short",
	}
	body, _ := json.Marshal(loginPayload)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", resp.StatusCode)
	}
}

func TestAuthLoginWrongPassword(t *testing.T) {
	loginPayload := model.LoginRequest{
		Email:    "admin@example.com",
		Password: "wrongpassword123",
	}
	body, _ := json.Marshal(loginPayload)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthLoginRateLimiter(t *testing.T) {
	defer middleware.LoginLimiter.ResetAll()

	loginPayload := model.LoginRequest{
		Email:    "ratelimit@example.com",
		Password: "wrongpassword123",
	}
	body, _ := json.Marshal(loginPayload)

	testIP := "198.51.100.99:5678"

	// Simulate > 5 failed attempts
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = testIP
		resp, _ := testApp.Test(req, -1)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Attempt %d: expected 401, got %d", i+1, resp.StatusCode)
		}
	}

	// 6th attempt should be 429
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = testIP
	resp, _ := testApp.Test(req, -1)
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 after 5 failed attempts, got %d", resp.StatusCode)
	}
}

func TestAuthMeSuccess(t *testing.T) {
	// First login as student
	loginPayload := model.LoginRequest{
		Email:    "rina.putri@student.siakad.ac.id",
		Password: "187221000001",
	}
	body, _ := json.Marshal(loginPayload)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	var loginResp struct {
		Data model.LoginResponseData `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp.Data.AccessToken

	// Test GET /auth/me
	meReq := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)

	meResp, err := testApp.Test(meReq, -1)
	if err != nil {
		t.Fatalf("Get me failed: %v", err)
	}

	if meResp.StatusCode != http.StatusOK {
		var errRes model.ApiResponse
		json.NewDecoder(meResp.Body).Decode(&errRes)
		t.Fatalf("Expected 200, got %d: %s (%+v)", meResp.StatusCode, errRes.Message, errRes.Errors)
	}

	var meData struct {
		Data model.AuthMeResponseData `json:"data"`
	}
	json.NewDecoder(meResp.Body).Decode(&meData)
	if meData.Data.Role != "mahasiswa" {
		t.Errorf("Expected role mahasiswa, got %s", meData.Data.Role)
	}
	if meData.Data.Student == nil || meData.Data.Student.NIM != "187221000001" {
		t.Errorf("Expected student profile with NIM 187221000001, got %+v", meData.Data.Student)
	}
}

func TestAuthMeWithoutToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	resp, _ := testApp.Test(req, -1)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 without token, got %d", resp.StatusCode)
	}
}
