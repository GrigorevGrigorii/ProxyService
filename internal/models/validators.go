package models

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	Validate *validator.Validate
	once     sync.Once
)

func init() {
	once.Do(func() {
		Validate = validator.New()

		Validate.RegisterValidation("duration", validDuration)
		Validate.RegisterValidation("query", validQuery)

		Validate.RegisterStructValidation(validateTarget, TargetDTO{})
	})
}

func validDuration(fl validator.FieldLevel) bool {
	duration := fl.Field().String()
	if _, err := time.ParseDuration(duration); err != nil {
		return false
	}
	return true
}

func validQuery(fl validator.FieldLevel) bool {
	query := fl.Field().String()
	if query == "" || query == "*" {
		return true
	}
	_, err := url.ParseQuery(query)
	if err != nil {
		return false
	}
	return true
}

func validateTarget(sl validator.StructLevel) {
	target := sl.Current().Interface().(TargetDTO)

	if target.CacheInterval != nil {
		if target.Method != http.MethodGet {
			sl.ReportError(
				target.CacheInterval,
				"CacheInterval",
				"cache_interval",
				"must_be_nil_for_non_get_methods",
				"'cache_interval' can be set only for GET method",
			)
		}
		if target.Query == "*" {
			sl.ReportError(
				target.CacheInterval,
				"CacheInterval",
				"cache_interval",
				"must_be_nil_for_*_query",
				"'cache_interval' can be set only non * query",
			)
		}
	}
}
