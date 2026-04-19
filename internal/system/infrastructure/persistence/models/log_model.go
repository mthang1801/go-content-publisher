package models

import "time"

type LogModel struct {
	ID        string    `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Level     string    `gorm:"column:level"`
	Module    string    `gorm:"column:module"`
	Message   string    `gorm:"column:message"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (LogModel) TableName() string {
	return "logs"
}
