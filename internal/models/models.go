package models

type TargetDTO struct {
	Path          string  `json:"path"`
	Method        string  `json:"method"`
	Query         string  `json:"query"`
	CacheInterval *string `json:"cache_interval"`
}

type ServiceDTO struct {
	Name          string      `json:"name"`
	Scheme        string      `json:"scheme"`
	Host          string      `json:"host"`
	Timeout       float32     `json:"timeout"`
	RetryCount    int         `json:"retry_count"`
	RetryInterval float32     `json:"retry_interval"`
	Version       int         `json:"version"`
	Targets       []TargetDTO `json:"targets"`
}
