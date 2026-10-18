package storage

import "time"

type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

type Entry struct {
	ServiceID  string
	OccurredAt time.Time
	ReceivedAt time.Time
	Message    string
	Level      Level
	Metadata   map[string]string
}
