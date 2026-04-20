package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// Time is a time.Time wrapper that scans SQLite TEXT timestamp columns.
// The modernc.org/sqlite driver does not auto-convert TEXT to time.Time
// (only true DATETIME-typed columns), so we parse on read and format on
// write here. JSON uses RFC3339 (same shape time.Time emits by default).
type Time struct {
	time.Time
}

func Now() Time {
	return Time{Time: time.Now().UTC()}
}

func NewTime(t time.Time) Time {
	return Time{Time: t}
}

// PtrNow returns a pointer to Now(), handy for nullable columns.
func PtrNow() *Time {
	t := Now()
	return &t
}

// sqlLayout is what we write back to SQLite: RFC3339Nano preserves
// sub-second precision and timezone, and is round-trippable by Scan.
const sqlLayout = time.RFC3339Nano

// scanLayouts covers every format we might read back:
//   - sqlLayout (our own writes going forward)
//   - Go's time.Time.String() default (what modernc writes when handed a time.Time)
//   - SQLite's datetime('now') default ("YYYY-MM-DD HH:MM:SS", used by column defaults)
var scanLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02 15:04:05.999999999 -0700 MST",
	"2006-01-02 15:04:05 -0700 MST",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05",
}

func (t *Time) Scan(src interface{}) error {
	if src == nil {
		t.Time = time.Time{}
		return nil
	}
	switch v := src.(type) {
	case time.Time:
		t.Time = v
		return nil
	case string:
		return t.parse(v)
	case []byte:
		return t.parse(string(v))
	default:
		return fmt.Errorf("models.Time: cannot scan %T", src)
	}
}

func (t Time) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	return t.Time.UTC().Format(sqlLayout), nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	return t.Time.MarshalJSON()
}

func (t *Time) UnmarshalJSON(data []byte) error {
	return t.Time.UnmarshalJSON(data)
}

func (t *Time) parse(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		t.Time = time.Time{}
		return nil
	}
	for _, layout := range scanLayouts {
		if parsed, err := time.Parse(layout, s); err == nil {
			t.Time = parsed
			return nil
		}
	}
	return fmt.Errorf("models.Time: no layout matched %q", s)
}
