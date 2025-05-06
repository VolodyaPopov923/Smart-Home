package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/gateways/http"
	"homework/internal/usecase"
	nethttp "net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

const (
	testTime     = "2025-04-21 12:47:30"
	testUsername = "VolodyaPopov923"
)

func getTestHistoryResponse(events []domain.Event) []http.SensorHistoryResponse {
	result := make([]http.SensorHistoryResponse, len(events))
	for i, e := range events {
		result[i] = http.SensorHistoryResponse{
			Timestamp:       e.Timestamp,
			Payload:         e.Payload,
			RequestTime:     testTime,
			RequestedByUser: testUsername,
		}
	}
	return result
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
	events, ok := args.Get(0).([]domain.Event)
	if !ok {
		return nil, errors.New("unexpected type assertion error")
	}
	return events, args.Error(1)
}

type MockSensorRepository struct {
	mock.Mock
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

func (m *MockSensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	args := m.Called(ctx, sn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	sensor, ok := args.Get(0).(*domain.Sensor)
	if !ok {
		return nil, errors.New("unexpected type assertion error")
	}
	return sensor, args.Error(1)
}

type MockWebSocketHandler struct{}

func (m *MockWebSocketHandler) Handle(_ *gin.Context, _ int64) error {
	return nil
}

func (m *MockWebSocketHandler) Shutdown() error {
	return nil
}

func TestSensorHistoryEndpoint(t *testing.T) {
	testCases := []struct {
		name           string
		sensorID       string
		startDate      string
		endDate        string
		acceptHeader   string
		setupMocks     func(*MockSensorRepository, *MockEventRepository)
		expectedStatus int
		expectedBody   []http.SensorHistoryResponse
		expectError    bool
	}{
		{
			name:         "Valid request with date range",
			sensorID:     "1",
			startDate:    "2025-01-01T00:00:00Z",
			endDate:      "2025-01-02T00:00:00Z",
			acceptHeader: "application/json",
			setupMocks: func(mockSensor *MockSensorRepository, mockEvent *MockEventRepository) {
				mockSensor.On("GetSensorByID", mock.Anything, int64(1)).Return(&domain.Sensor{ID: 1}, nil)

				events := []domain.Event{
					{SensorSerialNumber: "1234567890", Payload: 42, Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)},
					{SensorSerialNumber: "1234567890", Payload: 43, Timestamp: time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC)},
				}

				startDate, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
				endDate, _ := time.Parse(time.RFC3339, "2025-01-02T00:00:00Z")

				mockEvent.On("GetEventsHistoryBySensorID", mock.Anything, int64(1), startDate, endDate).Return(events, nil)
			},
			expectedStatus: nethttp.StatusOK,
			expectedBody: []http.SensorHistoryResponse{
				{Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC), Payload: 42, RequestTime: testTime, RequestedByUser: testUsername},
				{Timestamp: time.Date(2025, 1, 1, 13, 0, 0, 0, time.UTC), Payload: 43, RequestTime: testTime, RequestedByUser: testUsername},
			},
			expectError: false,
		},
		{
			name:         "Invalid sensor ID",
			sensorID:     "invalid",
			startDate:    "2025-01-01T00:00:00Z",
			endDate:      "2025-01-02T00:00:00Z",
			acceptHeader: "application/json",
			setupMocks: func(_ *MockSensorRepository, _ *MockEventRepository) {
			},
			expectedStatus: nethttp.StatusUnprocessableEntity,
			expectedBody:   nil,
			expectError:    true,
		},
		{
			name:         "Invalid start date format",
			sensorID:     "1",
			startDate:    "invalid-date",
			endDate:      "2025-01-02T00:00:00Z",
			acceptHeader: "application/json",
			setupMocks: func(_ *MockSensorRepository, _ *MockEventRepository) {
			},
			expectedStatus: nethttp.StatusUnprocessableEntity,
			expectedBody:   nil,
			expectError:    true,
		},
		{
			name:         "Invalid end date format",
			sensorID:     "1",
			startDate:    "2025-01-01T00:00:00Z",
			endDate:      "invalid-date",
			acceptHeader: "application/json",
			setupMocks: func(_ *MockSensorRepository, _ *MockEventRepository) {
			},
			expectedStatus: nethttp.StatusUnprocessableEntity,
			expectedBody:   nil,
			expectError:    true,
		},
		{
			name:         "Sensor not found",
			sensorID:     "999",
			startDate:    "2025-01-01T00:00:00Z",
			endDate:      "2025-01-02T00:00:00Z",
			acceptHeader: "application/json",
			setupMocks: func(mockSensor *MockSensorRepository, _ *MockEventRepository) {
				mockSensor.On("GetSensorByID", mock.Anything, int64(999)).Return(nil, usecase.ErrSensorNotFound)
			},
			expectedStatus: nethttp.StatusNotFound,
			expectedBody:   nil,
			expectError:    true,
		},
		{
			name:         "No date parameters (use defaults)",
			sensorID:     "1",
			startDate:    "",
			endDate:      "",
			acceptHeader: "application/json",
			setupMocks: func(mockSensor *MockSensorRepository, mockEvent *MockEventRepository) {
				mockSensor.On("GetSensorByID", mock.Anything, int64(1)).Return(&domain.Sensor{ID: 1}, nil)

				events := []domain.Event{
					{SensorSerialNumber: "1234567890", Payload: 42, Timestamp: time.Now().Add(-time.Hour)},
				}

				mockEvent.On("GetEventsHistoryBySensorID",
					mock.Anything,
					int64(1),
					mock.MatchedBy(func(_ time.Time) bool { return true }),
					mock.MatchedBy(func(_ time.Time) bool { return true }),
				).Return(events, nil)
			},
			expectedStatus: nethttp.StatusOK,
			expectedBody:   nil,
			expectError:    false,
		},
		{
			name:         "End date earlier than start date",
			sensorID:     "1",
			startDate:    "2025-01-02T00:00:00Z",
			endDate:      "2025-01-01T00:00:00Z",
			acceptHeader: "application/json",
			setupMocks: func(mockSensor *MockSensorRepository, mockEvent *MockEventRepository) {
				mockSensor.On("GetSensorByID", mock.Anything, int64(1)).Return(&domain.Sensor{ID: 1}, nil)

				startDate, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
				endDate, _ := time.Parse(time.RFC3339, "2025-01-02T00:00:00Z")

				mockEvent.On("GetEventsHistoryBySensorID",
					mock.Anything,
					int64(1),
					startDate,
					endDate,
				).Return([]domain.Event{}, nil)
			},
			expectedStatus: nethttp.StatusOK,
			expectedBody:   []http.SensorHistoryResponse{},
			expectError:    false,
		},
		{
			name:         "Database error",
			sensorID:     "1",
			startDate:    "2025-01-01T00:00:00Z",
			endDate:      "2025-01-02T00:00:00Z",
			acceptHeader: "application/json",
			setupMocks: func(mockSensor *MockSensorRepository, mockEvent *MockEventRepository) {
				mockSensor.On("GetSensorByID", mock.Anything, int64(1)).Return(&domain.Sensor{ID: 1}, nil)

				startDate, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
				endDate, _ := time.Parse(time.RFC3339, "2025-01-02T00:00:00Z")

				mockEvent.On("GetEventsHistoryBySensorID",
					mock.Anything,
					int64(1),
					startDate,
					endDate,
				).Return([]domain.Event{}, errors.New("database connection error"))
			},
			expectedStatus: nethttp.StatusInternalServerError,
			expectedBody:   nil,
			expectError:    true,
		},
		{
			name:         "Empty history",
			sensorID:     "1",
			startDate:    "2025-01-01T00:00:00Z",
			endDate:      "2025-01-02T00:00:00Z",
			acceptHeader: "application/json",
			setupMocks: func(mockSensor *MockSensorRepository, mockEvent *MockEventRepository) {
				mockSensor.On("GetSensorByID", mock.Anything, int64(1)).Return(&domain.Sensor{ID: 1}, nil)

				startDate, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
				endDate, _ := time.Parse(time.RFC3339, "2025-01-02T00:00:00Z")

				mockEvent.On("GetEventsHistoryBySensorID",
					mock.Anything,
					int64(1),
					startDate,
					endDate,
				).Return([]domain.Event{}, nil)
			},
			expectedStatus: nethttp.StatusOK,
			expectedBody:   []http.SensorHistoryResponse{},
			expectError:    false,
		},
		{
			name:         "Not acceptable format",
			sensorID:     "1",
			startDate:    "2025-01-01T00:00:00Z",
			endDate:      "2025-01-02T00:00:00Z",
			acceptHeader: "application/xml",
			setupMocks: func(_ *MockSensorRepository, _ *MockEventRepository) {
			},
			expectedStatus: nethttp.StatusNotAcceptable,
			expectedBody:   nil,
			expectError:    true,
		},
		{
			name:         "Any Accept header (*/*)",
			sensorID:     "1",
			startDate:    "2025-01-01T00:00:00Z",
			endDate:      "2025-01-02T00:00:00Z",
			acceptHeader: "*/*",
			setupMocks: func(mockSensor *MockSensorRepository, mockEvent *MockEventRepository) {
				mockSensor.On("GetSensorByID", mock.Anything, int64(1)).Return(&domain.Sensor{ID: 1}, nil)

				startDate, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
				endDate, _ := time.Parse(time.RFC3339, "2025-01-02T00:00:00Z")

				mockEvent.On("GetEventsHistoryBySensorID",
					mock.Anything,
					int64(1),
					startDate,
					endDate,
				).Return([]domain.Event{}, nil)
			},
			expectedStatus: nethttp.StatusOK,
			expectedBody:   []http.SensorHistoryResponse{},
			expectError:    false,
		},
		{
			name:         "Empty Accept header",
			sensorID:     "1",
			startDate:    "2025-01-01T00:00:00Z",
			endDate:      "2025-01-02T00:00:00Z",
			acceptHeader: "",
			setupMocks: func(mockSensor *MockSensorRepository, mockEvent *MockEventRepository) {
				mockSensor.On("GetSensorByID", mock.Anything, int64(1)).Return(&domain.Sensor{ID: 1}, nil)

				startDate, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
				endDate, _ := time.Parse(time.RFC3339, "2025-01-02T00:00:00Z")

				mockEvent.On("GetEventsHistoryBySensorID",
					mock.Anything,
					int64(1),
					startDate,
					endDate,
				).Return([]domain.Event{}, nil)
			},
			expectedStatus: nethttp.StatusOK,
			expectedBody:   []http.SensorHistoryResponse{},
			expectError:    false,
		},
		{
			name:         "Wrong sensor type error",
			sensorID:     "1",
			startDate:    "2025-01-01T00:00:00Z",
			endDate:      "2025-01-02T00:00:00Z",
			acceptHeader: "application/json",
			setupMocks: func(mockSensor *MockSensorRepository, mockEvent *MockEventRepository) {
				mockSensor.On("GetSensorByID", mock.Anything, int64(1)).Return(&domain.Sensor{ID: 1}, nil)

				startDate, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
				endDate, _ := time.Parse(time.RFC3339, "2025-01-02T00:00:00Z")

				mockEvent.On("GetEventsHistoryBySensorID",
					mock.Anything,
					int64(1),
					startDate,
					endDate,
				).Return([]domain.Event{}, usecase.ErrWrongSensorType)
			},
			expectedStatus: nethttp.StatusUnprocessableEntity,
			expectedBody:   nil,
			expectError:    true,
		},
		{
			name:         "Wrong sensor serial number error",
			sensorID:     "1",
			startDate:    "2025-01-01T00:00:00Z",
			endDate:      "2025-01-02T00:00:00Z",
			acceptHeader: "application/json",
			setupMocks: func(mockSensor *MockSensorRepository, mockEvent *MockEventRepository) {
				mockSensor.On("GetSensorByID", mock.Anything, int64(1)).Return(&domain.Sensor{ID: 1}, nil)

				startDate, _ := time.Parse(time.RFC3339, "2025-01-01T00:00:00Z")
				endDate, _ := time.Parse(time.RFC3339, "2025-01-02T00:00:00Z")

				mockEvent.On("GetEventsHistoryBySensorID",
					mock.Anything,
					int64(1),
					startDate,
					endDate,
				).Return([]domain.Event{}, usecase.ErrWrongSensorSerialNumber)
			},
			expectedStatus: nethttp.StatusUnprocessableEntity,
			expectedBody:   nil,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSensorRepo := new(MockSensorRepository)
			mockEventRepo := new(MockEventRepository)

			tc.setupMocks(mockSensorRepo, mockEventRepo)

			eventUsecase := &EventWithHistoryMock{
				mockEvent:  mockEventRepo,
				mockSensor: mockSensorRepo,
			}

			r := gin.New()

			sensorsGroup := r.Group("/sensors")

			sensorsGroup.GET("/:sensor_id/history", func(c *gin.Context) {
				if c.GetHeader("Accept") != "application/json" && c.GetHeader("Accept") != "" && !strings.Contains(c.GetHeader("Accept"), "*/*") {
					c.Status(nethttp.StatusNotAcceptable)
					return
				}

				id, err := strconv.ParseInt(c.Param("sensor_id"), 10, 64)
				if err != nil {
					c.JSON(nethttp.StatusUnprocessableEntity, http.ErrorResponse{Reason: "Invalid sensor ID"})
					return
				}

				startDateStr := c.Query("start_date")
				endDateStr := c.Query("end_date")

				var startDate, endDate time.Time
				var parseErr error

				if startDateStr != "" {
					startDate, parseErr = time.Parse(time.RFC3339, startDateStr)
					if parseErr != nil {
						c.JSON(nethttp.StatusUnprocessableEntity, http.ErrorResponse{Reason: "Invalid start_date format. Use RFC3339 format."})
						return
					}
				}

				if endDateStr != "" {
					endDate, parseErr = time.Parse(time.RFC3339, endDateStr)
					if parseErr != nil {
						c.JSON(nethttp.StatusUnprocessableEntity, http.ErrorResponse{Reason: "Invalid end_date format. Use RFC3339 format."})
						return
					}
				}

				if endDate.IsZero() {
					endDate = time.Now()
				}

				if startDate.IsZero() {
					startDate = endDate.AddDate(0, -1, 0)
				}

				if endDate.Before(startDate) {
					startDate, endDate = endDate, startDate
				}

				events, err := eventUsecase.GetSensorHistory(c.Request.Context(), id, startDate, endDate)
				if err != nil {
					switch {
					case errors.Is(err, usecase.ErrSensorNotFound):
						c.JSON(nethttp.StatusNotFound, http.ErrorResponse{Reason: err.Error()})
					case errors.Is(err, usecase.ErrWrongSensorSerialNumber) ||
						errors.Is(err, usecase.ErrWrongSensorType) ||
						errors.Is(err, usecase.ErrInvalidUserName) ||
						errors.Is(err, usecase.ErrInvalidEventTimestamp):
						c.JSON(nethttp.StatusUnprocessableEntity, http.ErrorResponse{Reason: err.Error()})
					default:
						c.JSON(nethttp.StatusInternalServerError, http.ErrorResponse{Reason: "Internal server error"})
					}
					return
				}

				c.JSON(nethttp.StatusOK, getTestHistoryResponse(events))
			})

			url := fmt.Sprintf("/sensors/%s/history", tc.sensorID)
			if tc.startDate != "" || tc.endDate != "" {
				url += "?"
				if tc.startDate != "" {
					url += "start_date=" + tc.startDate
				}
				if tc.startDate != "" && tc.endDate != "" {
					url += "&"
				}
				if tc.endDate != "" {
					url += "end_date=" + tc.endDate
				}
			}

			req, _ := nethttp.NewRequest("GET", url, nil)
			if tc.acceptHeader != "" {
				req.Header.Set("Accept", tc.acceptHeader)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if !tc.expectError && tc.expectedBody != nil {
				var response []http.SensorHistoryResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBody, response)
			}

			mockSensorRepo.AssertExpectations(t)
			mockEventRepo.AssertExpectations(t)
		})
	}
}

type EventWithHistoryMock struct {
	mockEvent  *MockEventRepository
	mockSensor *MockSensorRepository
}

func (e *EventWithHistoryMock) GetSensorHistory(ctx context.Context, id int64, startDate, endDate time.Time) ([]domain.Event, error) {
	sensor, err := e.mockSensor.GetSensorByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sensor == nil {
		return nil, usecase.ErrSensorNotFound
	}

	return e.mockEvent.GetEventsHistoryBySensorID(ctx, id, startDate, endDate)
}
