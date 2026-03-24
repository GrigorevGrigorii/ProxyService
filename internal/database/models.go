package database

type Target struct {
	ServiceName string `json:"-" gorm:"column:service_name;size:32;primaryKey"`
	Path        string `json:"path" gorm:"column:path;size:128;primaryKey;not null"`
	Method      string `json:"method" gorm:"column:method;type:httpmethod;primaryKey;not null"`

	Service Service `json:"-" gorm:"foreignKey:ServiceName;references:Name;constraint:OnDelete:CASCADE;"`
}

type Service struct {
	Name          string  `json:"name" gorm:"column:name;size:32;primaryKey"`
	Scheme        string  `json:"scheme" gorm:"column:scheme;type:httpscheme;not null"`
	Host          string  `json:"host" gorm:"column:host;size:128;unique;not null"`
	Timeout       float32 `json:"timeout" gorm:"column:timeout;type:numeric(4,2);not null"`
	RetryCount    int     `json:"retry_count" gorm:"column:retry_count;not null;default:0"`
	RetryInterval float32 `json:"retry_interval" gorm:"column:retry_interval;type:numeric(4,2);not null;default:0.0"`
	Version       int     `json:"version" gorm:"column:version;not null;default:0"`

	Targets []Target `json:"targets" gorm:"foreignKey:ServiceName;references:Name;constraint:OnDelete:CASCADE;"`
}

func (Service) TableName() string { return "services" }
func (Target) TableName() string  { return "targets" }
