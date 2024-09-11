package cache

import (
	"testing"
	"time"
)

func TestRecord_IsExpired(t *testing.T) {

	now := time.Now()
	record := &Record{created: now, expiry: DefaultCacheExpiry}

	if record.IsExpired() {
		t.Error("record should have 2 minutes till considered expired")
	}
}

func TestRecord_IsExpired2(t *testing.T) {

	fourMinutesAgo := time.Now().Add(-1 * 4 * time.Minute)
	record := &Record{created: fourMinutesAgo, expiry: DefaultCacheExpiry}

	if !record.IsExpired() {
		t.Error("expected record to be expired for 2 or more minutes")
	}
}

func TestRecord_IsExpired3(t *testing.T) {

	twoMinutesAgo := time.Now().Add(-1 * 2 * time.Minute)
	record := &Record{created: twoMinutesAgo, expiry: DefaultCacheExpiry}

	if !record.IsExpired() {
		t.Error("expected record to be expired exactly at this time")
	}
}
