package models

import (
	"net/url"
)

type TargetDTO struct {
	Path          string  `json:"path" validate:"required"`
	Method        string  `json:"method" validate:"required,oneof=GET POST PUT DELETE"`
	Query         string  `json:"query" validate:"query"`
	CacheInterval *string `json:"cache_interval" validate:"omitempty,duration"`
}

type ServiceDTO struct {
	Name          string      `json:"name" validate:"required,min=1"`
	Scheme        string      `json:"scheme" validate:"required,oneof=https http"`
	Host          string      `json:"host" validate:"required,min=1"`
	Timeout       float32     `json:"timeout" validate:"required,gt=0"`
	RetryCount    int         `json:"retry_count" validate:"gte=0"`
	RetryInterval float32     `json:"retry_interval" validate:"gte=0"`
	Version       int         `json:"version" validate:"gte=0"`
	Targets       []TargetDTO `json:"targets" validate:"dive"`
}

func (t *TargetDTO) SortQuery() error {
	if t.Query == "" || t.Query == "*" {
		return nil
	}
	parsedQuery, err := url.ParseQuery(t.Query)
	if err != nil {
		return err
	}
	t.Query = parsedQuery.Encode()
	return nil
}
