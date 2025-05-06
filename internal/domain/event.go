package domain

import "time"

// Event - структура события по датчику
type Event struct {
	// Timestamp - время события
	Timestamp time.Time
	// SensorSerialNumber - серийный номер датчика
	SensorSerialNumber string
	// SensorID - id датчика
	SensorID int64
	// Payload - данные события
	Payload int64
}
