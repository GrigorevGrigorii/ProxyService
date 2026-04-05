package models

import (
	"net/url"
)

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
