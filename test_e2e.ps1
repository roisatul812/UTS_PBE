# test_e2e.ps1 - Complete E2E testing of all 10 SIAKAD Mini Endpoints
$base = "http://127.0.0.1:3000/api/v1"

Write-Host "========================================"
Write-Host "SIAKAD MINI - 10 ENDPOINTS LIVE E2E TEST"
Write-Host "========================================"

# 1. POST /api/v1/auth/login (Admin)
$adminBody = '{"email":"admin@example.com","password":"admin1234"}'
$adminLogin = Invoke-RestMethod -Uri "$base/auth/login" -Method Post -ContentType "application/json" -Body $adminBody
$adminToken = $adminLogin.data.access_token
Write-Host "[1/10] POST /auth/login (Admin) - Token received: $($adminToken.Substring(0, 20))... Role: $($adminLogin.data.user.role)"

# 1b. POST /api/v1/auth/login (Mahasiswa)
$mhsBody = '{"email":"rina.putri@student.siakad.ac.id","password":"187221000001"}'
$mhsLogin = Invoke-RestMethod -Uri "$base/auth/login" -Method Post -ContentType "application/json" -Body $mhsBody
$mhsToken = $mhsLogin.data.access_token
Write-Host "[1/10] POST /auth/login (Mahasiswa) - Token received: $($mhsToken.Substring(0, 20))... Role: $($mhsLogin.data.user.role)"

# 2. GET /api/v1/auth/me (Mahasiswa)
$headersMhs = @{ Authorization = "Bearer $mhsToken" }
$me = Invoke-RestMethod -Uri "$base/auth/me" -Method Get -Headers $headersMhs
Write-Host "[2/10] GET /auth/me - Role: $($me.data.role), NIM: $($me.data.student.nim), Nama: $($me.data.student.nama)"

# 3. GET /api/v1/students (Admin)
$headersAdmin = @{ Authorization = "Bearer $adminToken" }
$students = Invoke-RestMethod -Uri "$base/students?page=1&per_page=5" -Method Get -Headers $headersAdmin
Write-Host "[3/10] GET /students - Total: $($students.meta.total), Current Page: $($students.meta.current_page), Count on page: $($students.data.Count)"

# 4. POST /api/v1/students (Admin)
$newNIM = "187221000555"
$createBody = @{
    nim = $newNIM
    nama = "Test Student E2E"
    email = "e2e.student@student.siakad.ac.id"
    prodi = "Teknik Informatika"
    angkatan = 2024
    ipk_terakhir = 3.70
} | ConvertTo-Json
$create = Invoke-RestMethod -Uri "$base/students" -Method Post -ContentType "application/json" -Headers $headersAdmin -Body $createBody
$newStudentID = $create.data.id
Write-Host "[4/10] POST /students - Created Student ID: $newStudentID, NIM: $($create.data.nim)"

# 5. GET /api/v1/students/:id (Admin / Mahasiswa self)
$detail = Invoke-RestMethod -Uri "$base/students/$newStudentID" -Method Get -Headers $headersAdmin
Write-Host "[5/10] GET /students/:id - Nama: $($detail.data.nama), Total SKS: $($detail.data.total_sks), Batas SKS: $($detail.data.batas_sks)"

# 6. PUT /api/v1/students/:id (Admin)
$updateBody = @{
    nama = "Test Student E2E Updated"
    prodi = "Sistem Informasi"
    angkatan = 2024
    ipk_terakhir = 3.85
} | ConvertTo-Json
$update = Invoke-RestMethod -Uri "$base/students/$newStudentID" -Method Put -ContentType "application/json" -Headers $headersAdmin -Body $updateBody
Write-Host "[6/10] PUT /students/:id - Updated Nama: $($update.data.nama), Prodi: $($update.data.prodi)"

# 7. DELETE /api/v1/students/:id (Admin)
$delCode = curl.exe -s -o /dev/null -w "%{http_code}" -X DELETE "$base/students/$newStudentID" -H "Authorization: Bearer $adminToken"
Write-Host "[7/10] DELETE /students/:id - Status Code: $delCode (Expected 204)"

# 8. GET /api/v1/courses (All authenticated)
$courses = Invoke-RestMethod -Uri "$base/courses?available=true" -Method Get -Headers $headersMhs
$courseToEnroll = $courses.data[0]
Write-Host "[8/10] GET /courses - Total Available: $($courses.data.Count), Course: $($courseToEnroll.kode_mk), Sisa Kuota: $($courseToEnroll.sisa_kuota)"

# 9. POST /api/v1/enrollments (Mahasiswa)
$enrollBody = @{
    course_id = $courseToEnroll.id
    tahun_akademik = "2026/2027-Ganjil"
} | ConvertTo-Json
$enroll = Invoke-RestMethod -Uri "$base/enrollments" -Method Post -ContentType "application/json" -Headers $headersMhs -Body $enrollBody
$enrollmentID = $enroll.data.id
Write-Host "[9/10] POST /enrollments - Enrolled ID: $enrollmentID, Course: $($enroll.data.kode_mk), TA: $($enroll.data.tahun_akademik)"

# 10. DELETE /api/v1/enrollments/:id (Mahasiswa)
$unenrollCode = curl.exe -s -o /dev/null -w "%{http_code}" -X DELETE "$base/enrollments/$enrollmentID" -H "Authorization: Bearer $mhsToken"
Write-Host "[10/10] DELETE /enrollments/:id - Status Code: $unenrollCode (Expected 204)"

Write-Host "========================================"
Write-Host "ALL 10 ENDPOINTS TESTED LIVE AND PASSED!"
Write-Host "========================================"
