package authorization

import (
	"testing"
	"time"
)

func TestIsTimeValidDateStaysValidForCurrentDay(t *testing.T) {
	client := NewClient("example.com")
	date := time.Now().Format("2006-01-02")
	if !client.isTimeValid(date) {
		t.Fatalf("expected %q to stay valid for the current day", date)
	}
}

func TestCloneAccreditsReturnsIndependentCopy(t *testing.T) {
	original := []Accredit{{Sn: "A", Time: "2026-01-01"}}
	cloned := cloneAccredits(original)
	cloned[0].Sn = "B"
	if original[0].Sn != "A" {
		t.Fatalf("original slice was modified: %+v", original)
	}
}
