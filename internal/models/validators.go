package models

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

func (t *TargetDTO) Validate() error {
	// Check service.Targets[i].Query
	query := t.Query
	if len(query) > 0 && query != "*" {
		parsedQuery, err := url.ParseQuery(t.Query)
		if err != nil {
			return err
		}
		t.Query = parsedQuery.Encode()
	}

	// Check service.Targets[i].CacheInterval
	if t.CacheInterval != nil {
		if t.Method != http.MethodGet {
			return errors.New("'cache_interval' can be set only for GET requests")
		}
		if t.Query == "*" {
			return errors.New("Not allowed to set cache interval for target with query='*'")
		}
		if _, err := time.ParseDuration(*t.CacheInterval); err != nil {
			return err
		}
	}

	return nil
}

func (s *ServiceDTO) Validate() error {
	for i := range s.Targets {
		err := s.Targets[i].Validate()
		if err != nil {
			return err
		}
	}
	return nil
}
