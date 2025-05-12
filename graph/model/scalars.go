package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Implement Scanner/Valuer for custom enums
func (w *Weekday) Scan(value interface{}) error {
	if s, ok := value.(string); ok {
		*w = Weekday(s)
		return nil
	}
	return errors.New("invalid weekday value")
}

func (w Weekday) Value() (driver.Value, error) {
	return string(w), nil
}

// Implement similar for other enums if needed
// ...

// Custom JSON handling for Weekdays
func (w *Weekdays) UnmarshalJSON(data []byte) error {
	var days []Weekday
	if err := json.Unmarshal(data, &days); err != nil {
		return err
	}
	*w = days
	return nil
}

func (w Weekdays) MarshalJSON() ([]byte, error) {
	return json.Marshal([]Weekday(w))
}
