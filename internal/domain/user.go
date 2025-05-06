package domain

// User - структура для хранения пользователя
type User struct {
	// ID - id пользователя
	ID int64
	// Name - имя пользователя
	Name string
}

// SensorOwner - структура для связи пользователя и датчика
// Связь многие-ко-многим: пользователь может иметь доступ к нескольким датчикам, датчик может быть доступен для нескольких пользователей.
type SensorOwner struct {
	// UserID - id пользователя
	UserID int64
	// SensorID - id датчика
	SensorID int64
}
