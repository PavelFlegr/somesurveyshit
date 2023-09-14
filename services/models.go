package services

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Survey struct {
	Id        int64
	Title     string
	Questions []Question
	Created   time.Time
	Updated   time.Time
}

type Question struct {
	Id          int64
	Title       string
	Description string
	Options     OptionsSlice
}

type Option struct {
	Label string `json:label`
}

type OptionsSlice []Option

func (option OptionsSlice) Value() (driver.Value, error) {
	return json.Marshal(option)
}

func (option *OptionsSlice) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &option)
}
