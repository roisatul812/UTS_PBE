-- Migration: 004_create_enrollments.sql
CREATE TABLE IF NOT EXISTS enrollments (
    id SERIAL PRIMARY KEY,
    student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id INT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    tahun_akademik VARCHAR(30) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_student_course_semester UNIQUE (student_id, course_id, tahun_akademik)
);

CREATE INDEX IF NOT EXISTS idx_enrollments_student_ta ON enrollments(student_id, tahun_akademik);
CREATE INDEX IF NOT EXISTS idx_enrollments_course_ta ON enrollments(course_id, tahun_akademik);
CREATE INDEX IF NOT EXISTS idx_enrollments_course ON enrollments(course_id);
