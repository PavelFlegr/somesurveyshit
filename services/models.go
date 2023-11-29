package services

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Entity interface {
	getId() int64
}

type TemplateData struct {
	LoggedIn bool
	Data     interface{}
}

type User struct {
	Id       int64
	Email    string
	Password string
}

type Survey struct {
	Id      int64
	Title   string
	Blocks  []Block
	Created time.Time
	Updated time.Time
}

type Block struct {
	Id        int64
	Title     string
	SurveyId  int64
	Randomize bool
	Questions []Question
}

type Question struct {
	Id            int64
	Title         string
	Description   string
	SurveyId      int64
	BlockId       int64
	Configuration Configuration
}

type Configuration struct {
	Options      []Option `json:"options"`
	Randomize    bool     `json:"randomize"`
	QuestionType string   `json:"questionType"`
}

type Option struct {
	Id    int64
	Label string `json:"label"`
}

func (configuration Configuration) Value() (driver.Value, error) {
	return json.Marshal(configuration)
}

func (configuration *Configuration) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &configuration)
}
