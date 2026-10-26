package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cartabinaria/unibo-go/curriculum"
	"github.com/cartabinaria/unibo-go/exams"
	"github.com/cartabinaria/unibo-go/timetable"

	"github.com/VaiTon/unibocalendar/unibo_integ"
)

type fakeCourseAPI struct {
	curricula map[int]curriculum.Curricula
	timetable timetable.Timetable
}

func (f fakeCourseAPI) GetCourseWebsiteID(*unibo_integ.Course) (unibo_integ.CourseId, error) {
	return unibo_integ.CourseId{Tipologia: "laurea", Id: "test-course"}, nil
}

func (f fakeCourseAPI) GetAllCurricula(*unibo_integ.Course) (map[int]curriculum.Curricula, error) {
	return f.curricula, nil
}

func (f fakeCourseAPI) GetTimetable(*unibo_integ.Course, int, curriculum.Curriculum, *timetable.Interval) (timetable.Timetable, error) {
	return f.timetable, nil
}

func (fakeCourseAPI) GetExams(string, string) ([]exams.Exam, error) {
	return nil, nil
}

func TestCoursePageRendersFakeCourseData(t *testing.T) {
	curr := curriculum.Curriculum{Value: "general", Label: "General"}
	course := unibo_integ.Course{
		Codice:      900001,
		Descrizione: "Test course",
		Tipologia:   "Laurea",
		DurataAnni:  1,
		Url:         "https://example.test/course",
	}
	api := fakeCourseAPI{
		curricula: map[int]curriculum.Curricula{1: {curr}},
		timetable: timetable.Timetable{{CodModulo: "SUB001", Title: "Test subject"}},
	}
	router := setupRouterWithAPI(unibo_integ.CoursesMap{course.Codice: course}, api)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/courses/900001", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body.String())
	}
	for _, want := range []string{"Test course", "Test subject", "/cal/900001/1?curr=general"} {
		if !strings.Contains(response.Body.String(), want) {
			t.Errorf("response body does not contain %q", want)
		}
	}
}

func TestCourseCalendarUsesFakeTimetable(t *testing.T) {
	start := time.Date(2026, time.October, 26, 9, 0, 0, 0, time.Local)
	course := unibo_integ.Course{Codice: 900002, Descrizione: "Calendar test", DurataAnni: 1}
	api := fakeCourseAPI{
		timetable: timetable.Timetable{{
			CodModulo: "SUB002",
			Title:     "Calendar subject",
			Teacher:   "Test teacher",
			Start:     timetable.CalendarTime{Time: start},
			End:       timetable.CalendarTime{Time: start.Add(time.Hour)},
		}},
	}
	router := setupRouterWithAPI(unibo_integ.CoursesMap{course.Codice: course}, api)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/cal/900002/1", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "text/calendar; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", got, "text/calendar; charset=utf-8")
	}
	for _, want := range []string{"BEGIN:VCALENDAR", "Calendar subject", "END:VCALENDAR"} {
		if !strings.Contains(response.Body.String(), want) {
			t.Errorf("calendar body does not contain %q", want)
		}
	}
}

func TestRouterRejectsInvalidAndUnknownCourseRequests(t *testing.T) {
	router := setupRouter(make(unibo_integ.CoursesMap))

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "course id is not a number",
			path:       "/courses/not-a-number",
			wantStatus: http.StatusBadRequest,
			wantBody:   "Invalid course id",
		},
		{
			name:       "course does not exist",
			path:       "/courses/123",
			wantStatus: http.StatusNotFound,
			wantBody:   "Course not found",
		},
		{
			name:       "calendar year is not a number",
			path:       "/cal/123/not-a-number",
			wantStatus: http.StatusBadRequest,
			wantBody:   "Invalid year",
		},
		{
			name:       "calendar course id is not a number",
			path:       "/cal/not-a-number/1",
			wantStatus: http.StatusBadRequest,
			wantBody:   "Invalid id",
		},
		{
			name:       "calendar course does not exist",
			path:       "/cal/123/1",
			wantStatus: http.StatusNotFound,
			wantBody:   "Course not found",
		},
		{
			name:       "exam year is not a number",
			path:       "/exams/123/not-a-number",
			wantStatus: http.StatusBadRequest,
			wantBody:   "Invalid year",
		},
		{
			name:       "exam course id is not a number",
			path:       "/exams/not-a-number/1",
			wantStatus: http.StatusBadRequest,
			wantBody:   "Invalid id",
		},
		{
			name:       "exam course does not exist",
			path:       "/exams/123/1",
			wantStatus: http.StatusNotFound,
			wantBody:   "Course not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)

			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if body := response.Body.String(); body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}
