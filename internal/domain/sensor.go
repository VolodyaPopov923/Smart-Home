package domain

import "time"

type SensorType string

const (
	SensorTypeContactClosure SensorType = "cc"
	SensorTypeADC            SensorType = "adc"
)

// Sensor - структура для хранения данных датчика
type Sensor struct {
	// ID - id датчика
	ID int64
	// SerialNumber - серийный номер датчика
	SerialNumber string
	// Type - тип датчика
	Type SensorType
	// CurrentState - текущее состояние датчика
	CurrentState int64
	// Description - описание датчика
	Description string
	// IsActive - активен ли датчик
	IsActive bool
	// RegisteredAt - дата регистрации датчика
	RegisteredAt time.Time
	// LastActivity - дата последнего изменения состояния датчика
	LastActivity time.Time
}
