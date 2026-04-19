package models

type SettingModel struct {
	Key         string  `gorm:"column:key;primaryKey"`
	Value       string  `gorm:"column:value"`
	JSONValue   []byte  `gorm:"column:json_value;type:jsonb"`
	Description *string `gorm:"column:description"`
}

func (SettingModel) TableName() string {
	return "settings"
}
