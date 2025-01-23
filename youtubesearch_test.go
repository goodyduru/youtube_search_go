package youtubesearch

import (
	"testing"
	"time"
)

func TestNoTimeout(t *testing.T) {
	query := "Rob Pike Go speech"
	results, _ := Search(query, 0)
	if results == nil {
		t.Errorf("got a nil result")
	}
}

func TestWithLargeEnoughTimeout(t *testing.T) {
	query := "Rob Pike Go speech"
	timeout := time.Duration(10_000_000_000) // 10 seconds
	results, _ := Search(query, timeout)
	if results == nil {
		t.Errorf("got a nil result")
	}
}

func TestWithSmallTimeout(t *testing.T) {
	query := "Rob Pike Go speech"
	timeout := time.Duration(10_000) // 10 microseconds
	_, err := Search(query, timeout)
	if err == nil {
		t.Errorf("a timeout error should be returned")
	}
}
