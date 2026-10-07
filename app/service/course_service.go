package service

import (
	"uts-pbe-siakad/app/model"
	"uts-pbe-siakad/app/repository"
)

type CourseService interface {
	GetAll(q model.CourseListQuery) ([]model.Course, error)
}

type courseService struct {
	courseRepo repository.CourseRepository
}

func NewCourseService(courseRepo repository.CourseRepository) CourseService {
	return &courseService{courseRepo: courseRepo}
}

func (s *courseService) GetAll(q model.CourseListQuery) ([]model.Course, error) {
	return s.courseRepo.FindAll(q)
}
