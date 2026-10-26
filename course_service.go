package main

import (
	"github.com/cartabinaria/unibo-go/curriculum"
	"github.com/cartabinaria/unibo-go/exams"
	"github.com/cartabinaria/unibo-go/timetable"

	"github.com/VaiTon/unibocalendar/unibo_integ"
)

// courseAPI isolates handlers from the university API so their behavior can
// be tested with predictable in-memory responses.
type courseAPI interface {
	GetCourseWebsiteID(course *unibo_integ.Course) (unibo_integ.CourseId, error)
	GetAllCurricula(course *unibo_integ.Course) (map[int]curriculum.Curricula, error)
	GetTimetable(course *unibo_integ.Course, year int, curr curriculum.Curriculum, period *timetable.Interval) (timetable.Timetable, error)
	GetExams(courseType, courseID string) ([]exams.Exam, error)
}

type application struct {
	courses unibo_integ.CoursesMap
	api     courseAPI
}

type liveCourseAPI struct{}

func (liveCourseAPI) GetCourseWebsiteID(course *unibo_integ.Course) (unibo_integ.CourseId, error) {
	return course.GetCourseWebsiteId()
}

func (liveCourseAPI) GetAllCurricula(course *unibo_integ.Course) (map[int]curriculum.Curricula, error) {
	return course.GetAllCurricula()
}

func (liveCourseAPI) GetTimetable(course *unibo_integ.Course, year int, curr curriculum.Curriculum, period *timetable.Interval) (timetable.Timetable, error) {
	return course.GetTimetable(year, curr, period)
}

func (liveCourseAPI) GetExams(courseType, courseID string) ([]exams.Exam, error) {
	return exams.GetExams(courseType, courseID)
}
