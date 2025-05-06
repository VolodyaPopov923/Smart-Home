package http

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/usecase"
	nethttp "net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

const (
	fuzzTestTime     = "2025-04-21 17:23:23"
	fuzzTestUsername = "VolodyaPopov923"
)

type MockSensorRepository struct {
	mock.Mock
}

func (m *MockSensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	sensor, ok := args.Get(0).(*domain.Sensor)
	if !ok {
		return nil, errors.New("unexpected type assertion error")
	}
	return sensor, args.Error(1)
}

func (m *MockSensorRepository) GetSensorBySerialNumber(ctx context.Context, serialNumber string) (*domain.Sensor, error) {
	args := m.Called(ctx, serialNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	sensor, ok := args.Get(0).(*domain.Sensor)
	if !ok {
		return nil, errors.New("unexpected type assertion error")
	}
	return sensor, args.Error(1)
}

func (m *MockSensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	args := m.Called(ctx, sensor)
	return args.Error(0)
}

func (m *MockSensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	args := m.Called(ctx)
	sensors, ok := args.Get(0).([]domain.Sensor)
	if !ok {
		return nil, errors.New("unexpected type assertion error")
	}
	return sensors, args.Error(1)
}

type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	event, ok := args.Get(0).(*domain.Event)
	if !ok {
		return nil, errors.New("unexpected type assertion error")
	}
	return event, args.Error(1)
}

func (m *MockEventRepository) GetEventsHistoryBySensorID(ctx context.Context, id int64, startDate, endDate time.Time) ([]domain.Event, error) {
	args := m.Called(ctx, id, startDate, endDate)
	if args.Get(0) == nil {
		return []domain.Event{}, args.Error(1)
	}
	events, ok := args.Get(0).([]domain.Event)
	if !ok {
		return []domain.Event{}, errors.New("unexpected type assertion error")
	}
	return events, args.Error(1)
}

type EventWithHistoryMock struct {
	mockEvent  *MockEventRepository
	mockSensor *MockSensorRepository
}

func (e *EventWithHistoryMock) GetSensorHistory(ctx context.Context, sensorID int64, startDate, endDate time.Time) ([]domain.Event, error) {
	if e.mockSensor == nil || e.mockEvent == nil {
		return []domain.Event{}, errors.New("mock repositories not initialized")
	}

	sensor, err := e.mockSensor.GetSensorByID(ctx, sensorID)
	if err != nil {
		return []domain.Event{}, err
	}

	if sensor == nil {
		return []domain.Event{}, usecase.ErrSensorNotFound
	}

	events, err := e.mockEvent.GetEventsHistoryBySensorID(ctx, sensorID, startDate, endDate)
	if err != nil {
		return []domain.Event{}, err
	}

	return events, nil
}

func FuzzSensorHistoryDateParsing(f *testing.F) {
	f.Add("2025-01-01T00:00:00Z", "2025-01-02T00:00:00Z")
	f.Add("2025-01-01T00:00:00+00:00", "2025-01-02T00:00:00-00:00")
	f.Add("", "")
	f.Add("2025-01-01", "2025-01-02")
	f.Add("2025-01-02T00:00:00Z", "2025-01-01T00:00:00Z")
	f.Add("2025-13-01T00:00:00Z", "2025-01-02T00:00:00Z")
	f.Add("2025-01-01T00:00:00Z", "2025-13-02T00:00:00Z")
	f.Add("2025-01-01T25:00:00Z", "2025-01-02T00:00:00Z")
	f.Add("2025-01-32T00:00:00Z", "2025-01-02T00:00:00Z")
	f.Add("2025-01-01T00:00:00Z", "2025-01-32T00:00:00Z")
	f.Add("2025-01-01T00:60:00Z", "2025-01-02T00:00:00Z")
	f.Add("2025-01-01T00:00:60Z", "2025-01-02T00:00:00Z")

	f.Fuzz(func(t *testing.T, startDate, endDate string) {
		mockSensorRepo := new(MockSensorRepository)
		mockEventRepo := new(MockEventRepository)

		mockSensorRepo.On("GetSensorByID", mock.Anything, int64(1)).Return(&domain.Sensor{ID: 1}, nil)
		mockEventRepo.On("GetEventsHistoryBySensorID",
			mock.Anything,
			int64(1),
			mock.MatchedBy(func(_ time.Time) bool { return true }),
			mock.MatchedBy(func(_ time.Time) bool { return true }),
		).Return([]domain.Event{}, nil)

		r := gin.New()

		eventUsecase := &EventWithHistoryMock{
			mockEvent:  mockEventRepo,
			mockSensor: mockSensorRepo,
		}

		r.GET("/sensors/:sensor_id/history", func(c *gin.Context) {
			if c.GetHeader("Accept") != "application/json" && c.GetHeader("Accept") != "" && !strings.Contains(c.GetHeader("Accept"), "*/*") {
				c.Status(nethttp.StatusNotAcceptable)
				return
			}

			id, err := strconv.ParseInt(c.Param("sensor_id"), 10, 64)
			if err != nil {
				c.JSON(nethttp.StatusUnprocessableEntity, ErrorResponse{Reason: "Invalid sensor ID"})
				return
			}

			startDateStr := c.Query("start_date")
			endDateStr := c.Query("end_date")

			var startDateParsed, endDateParsed time.Time
			var parseErr error

			if startDateStr != "" {
				startDateParsed, parseErr = time.Parse(time.RFC3339, startDateStr)
				if parseErr != nil {
					c.JSON(nethttp.StatusUnprocessableEntity, ErrorResponse{Reason: "Invalid start_date format. Use RFC3339 format."})
					return
				}
			}

			if endDateStr != "" {
				endDateParsed, parseErr = time.Parse(time.RFC3339, endDateStr)
				if parseErr != nil {
					c.JSON(nethttp.StatusUnprocessableEntity, ErrorResponse{Reason: "Invalid end_date format. Use RFC3339 format."})
					return
				}
			}

			if endDateParsed.IsZero() {
				endDateParsed = time.Now()
			}

			if startDateParsed.IsZero() {
				startDateParsed = endDateParsed.AddDate(0, -1, 0)
			}

			if endDateParsed.Before(startDateParsed) {
				startDateParsed, endDateParsed = endDateParsed, startDateParsed
			}

			if eventUsecase == nil {
				c.JSON(nethttp.StatusInternalServerError, ErrorResponse{Reason: "Server configuration error"})
				return
			}

			events, err := eventUsecase.GetSensorHistory(c.Request.Context(), id, startDateParsed, endDateParsed)
			if err != nil {
				switch {
				case errors.Is(err, usecase.ErrSensorNotFound):
					c.JSON(nethttp.StatusNotFound, ErrorResponse{Reason: err.Error()})
				case errors.Is(err, usecase.ErrWrongSensorSerialNumber) ||
					errors.Is(err, usecase.ErrWrongSensorType) ||
					errors.Is(err, usecase.ErrInvalidUserName) ||
					errors.Is(err, usecase.ErrInvalidEventTimestamp):
					c.JSON(nethttp.StatusUnprocessableEntity, ErrorResponse{Reason: err.Error()})
				default:
					c.JSON(nethttp.StatusInternalServerError, ErrorResponse{Reason: "Internal server error"})
				}
				return
			}

			if events == nil {
				events = []domain.Event{}
			}

			result := make([]SensorHistoryResponse, len(events))
			for i, e := range events {
				result[i] = SensorHistoryResponse{
					Timestamp:       e.Timestamp,
					Payload:         e.Payload,
					RequestTime:     fuzzTestTime,
					RequestedByUser: fuzzTestUsername,
				}
			}

			c.JSON(nethttp.StatusOK, result)
		})

		acceptHeaders := []string{"application/json", "", "*/*", "application/xml"}

		for _, acceptHeader := range acceptHeaders {
			escapedStartDate := url.QueryEscape(startDate)
			escapedEndDate := url.QueryEscape(endDate)

			urlPath := fmt.Sprintf("/sensors/1/history?start_date=%s&end_date=%s", escapedStartDate, escapedEndDate)

			req, err := nethttp.NewRequest("GET", urlPath, nil)
			if err != nil {
				t.Logf("Skipping invalid URL: %v", err)
				continue
			}

			if acceptHeader != "" {
				req.Header.Set("Accept", acceptHeader)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			status := w.Code

			if acceptHeader == "application/xml" && status != nethttp.StatusNotAcceptable {
				t.Errorf("Expected 406 Not Acceptable for Accept: %s, got %d", acceptHeader, status)
			} else if acceptHeader != "application/xml" && status != nethttp.StatusOK && status != nethttp.StatusUnprocessableEntity {
				t.Errorf("Unexpected status code %d for date params start_date=%s, end_date=%s, Accept: %s",
					status, startDate, endDate, acceptHeader)
			}
		}
	})
}
