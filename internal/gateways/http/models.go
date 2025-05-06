package http

import (
	"homework/internal/domain"
	"time"
)

type SensorEventRequest struct {
	SensorSerialNumber string `json:"sensor_serial_number"`
	Payload            int64  `json:"payload"`
}

type SensorCreateRequest struct {
	SerialNumber string `json:"serial_number"`
	Type         string `json:"type"`
	Description  string `json:"description"`
	IsActive     bool   `json:"is_active"`
}

type UserCreateRequest struct {
	Name string `json:"name"`
}

type SensorBindingRequest struct {
	SensorID int64 `json:"sensor_id"`
}

type ErrorResponse struct {
	Reason string `json:"reason"`
}

type SensorResponse struct {
	ID           int64     `json:"id"`
	SerialNumber string    `json:"serial_number"`
	Type         string    `json:"type"`
	CurrentState int64     `json:"current_state"`
	Description  string    `json:"description"`
	IsActive     bool      `json:"is_active"`
	RegisteredAt time.Time `json:"registered_at"`
	LastActivity time.Time `json:"last_activity"`
}

type UserResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type SensorHistoryResponse struct {
	Timestamp       time.Time `json:"timestamp"`
	Payload         int64     `json:"payload"`
	RequestTime     string    `json:"request_time"`
	RequestedByUser string    `json:"requested_by_user"`
}

type SensorHistoryMetadata struct {
	RequestTime     string
	RequestedByUser string
}

func sensorToDomain(req SensorCreateRequest) *domain.Sensor {
	return &domain.Sensor{
		SerialNumber: req.SerialNumber,
		Type:         domain.SensorType(req.Type),
		Description:  req.Description,
		IsActive:     req.IsActive,
		RegisteredAt: time.Now(),
		LastActivity: time.Now(),
	}
}

func sensorToResponse(s *domain.Sensor) SensorResponse {
	return SensorResponse{
		ID:           s.ID,
		SerialNumber: s.SerialNumber,
		Type:         string(s.Type),
		CurrentState: s.CurrentState,
		Description:  s.Description,
		IsActive:     s.IsActive,
		RegisteredAt: s.RegisteredAt,
		LastActivity: s.LastActivity,
	}
}

func sensorsToResponse(sensors []domain.Sensor) []SensorResponse {
	result := make([]SensorResponse, len(sensors))
	for i, s := range sensors {
		result[i] = sensorToResponse(&s)
	}
	return result
}

func userToDomain(req UserCreateRequest) *domain.User {
	return &domain.User{
		Name: req.Name,
	}
}

func userToResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:   u.ID,
		Name: u.Name,
	}
}

func eventToDomain(req SensorEventRequest) *domain.Event {
	return &domain.Event{
		SensorSerialNumber: req.SensorSerialNumber,
		Payload:            req.Payload,
		Timestamp:          time.Now(),
	}
}

func eventsToHistoryResponse(events []domain.Event, metadata SensorHistoryMetadata) []SensorHistoryResponse {
	result := make([]SensorHistoryResponse, len(events))

	for i, e := range events {
		result[i] = SensorHistoryResponse{
			Timestamp:       e.Timestamp,
			Payload:         e.Payload,
			RequestTime:     metadata.RequestTime,
			RequestedByUser: metadata.RequestedByUser,
		}
	}
	return result
}
