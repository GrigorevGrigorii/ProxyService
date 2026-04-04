package models

import (
	"net/http"
	"testing"
)

func TestSortQuery(t *testing.T) {
	target := TargetDTO{Path: "/", Method: http.MethodGet, Query: "b=c&a=b", CacheInterval: nil}
	err := target.SortQuery()

	if err != nil {
		t.Fatalf("Got not nil error: %s", err.Error())
	}
	if target.Query != "a=b&b=c" {
		t.Fatalf("Expected query 'a=b&b=c', got '%s'", target.Query)
	}
}
