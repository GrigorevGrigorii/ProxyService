package handlers

type target struct {
	Path          *string `json:"path" validate:"required"`
	Method        *string `json:"method" validate:"required,oneof=GET POST PUT DELETE"`
	Query         *string `json:"query" validate:"required,query"`
	CacheInterval *string `json:"cache_interval" validate:"omitempty,duration"`
}

type createServiceRequest struct {
	Name          *string  `json:"name" validate:"required,min=1"`
	Scheme        *string  `json:"scheme" validate:"required,oneof=https http"`
	Host          *string  `json:"host" validate:"required,min=1"`
	Timeout       *float32 `json:"timeout" validate:"required,gt=0"`
	RetryCount    *int     `json:"retry_count" validate:"required,gte=0"`
	RetryInterval *float32 `json:"retry_interval" validate:"required,gte=0"`
	Targets       []target `json:"targets" validate:"required,dive"`
}

type updateServiceRequest struct {
	Scheme        *string  `json:"scheme" validate:"required,oneof=https http"`
	Host          *string  `json:"host" validate:"required,min=1"`
	Timeout       *float32 `json:"timeout" validate:"required,gt=0"`
	RetryCount    *int     `json:"retry_count" validate:"required,gte=0"`
	RetryInterval *float32 `json:"retry_interval" validate:"required,gte=0"`
	Version       *int     `json:"version" validate:"required,gte=0"`
	Targets       []target `json:"targets" validate:"required,dive"`
}
