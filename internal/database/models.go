package database

type Target struct {
	ServiceName   string `gorm:"column:service_name;size:32;primaryKey"`
	Path          string `gorm:"column:path;size:128;primaryKey;not null"`
	Method        string `gorm:"column:method;type:httpmethod;primaryKey;not null"`
	Query         string `gorm:"column:query;type:text;primaryKey;not null;default:''"`
	CacheInterval *int   `gorm:"column:cache_interval"`

	Service Service `gorm:"foreignKey:ServiceName;references:Name;constraint:OnDelete:CASCADE;"`
}

type Service struct {
	Name          string  `gorm:"column:name;size:32;primaryKey"`
	Scheme        string  `gorm:"column:scheme;type:httpscheme;not null"`
	Host          string  `gorm:"column:host;size:128;unique;not null"`
	Timeout       float32 `gorm:"column:timeout;type:numeric(4,2);not null"`
	RetryCount    int     `gorm:"column:retry_count;not null;default:0"`
	RetryInterval float32 `gorm:"column:retry_interval;type:numeric(4,2);not null;default:0.0"`
	Version       int     `gorm:"column:version;not null;default:0"`

	Targets []Target `gorm:"foreignKey:ServiceName;references:Name;constraint:OnDelete:CASCADE;"`
}

func (Service) TableName() string { return "services" }
func (Target) TableName() string  { return "targets" }
