package services

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/lib/pq"
	"log"
	"main/context"
	"time"
)

func ListSurveys(userId int64) []Survey {
	rows, err := context.Ctx.Db.Query("select id, title, created, updated from surveys inner join public.survey_permissions sp on surveys.id = sp.entity_id where sp.user_id = $1 and sp.action = 'read'", userId)
	if err != nil {
		log.Panic("Couldn't list surveys", err)
	}
	surveys := []Survey{}
	for rows.Next() {
		var survey Survey
		err = rows.Scan(&survey.Id, &survey.Title, &survey.Created, &survey.Updated)
		if err != nil {
			log.Panic("Couldn't scan survey", err)
		}
		surveys = append(surveys, survey)
	}

	return surveys
}

func GetSurvey(surveyId int64) Survey {
	var survey Survey
	survey.Id = surveyId
	var created string
	var updated string
	var questionsOrder []int64
	err := context.Ctx.Db.QueryRow("select title, created, updated, questions_order from surveys where id = $1", surveyId).Scan(&survey.Title, &created, &updated, (*pq.Int64Array)(&questionsOrder))
	if err != nil {
		log.Print("Couldn't get survey", err)
	}

	createdTime, _ := time.Parse(time.RFC3339, created)
	updatedTime, _ := time.Parse(time.RFC3339, updated)
	survey.Created = createdTime
	survey.Updated = updatedTime

	tmp := make(map[int64]Question)
	survey.Questions = ListQuestions(surveyId)
	for _, v := range survey.Questions {
		tmp[v.Id] = v
	}
	for i, v := range questionsOrder {
		survey.Questions[i] = tmp[v]
	}

	return survey
}

func CreateSurvey(title string, userId int64) Survey {
	var survey Survey
	survey.Title = title
	err := context.Ctx.Db.QueryRow("insert into surveys (title, created, updated, user_id) values ($1, now(), now(), $2) returning id, created, updated", title, userId).Scan(&survey.Id, &survey.Created, &survey.Updated)
	if err != nil {
		log.Print("Couldn't create survey", err)
	}

	AddPermission(userId, "survey", survey.Id, "manage")
	AddPermission(userId, "survey", survey.Id, "edit")
	AddPermission(userId, "survey", survey.Id, "read")

	return survey
}

func RenameSurvey(surveyId int64, title string) {
	_, err := context.Ctx.Db.Exec("update surveys set updated = now(), title = $1 where id = $2", title, surveyId)
	if err != nil {
		log.Print("Couldn't update survey title", err)
	}
}

func DeleteSurvey(surveyId int64, userId int64) {
	_, err := context.Ctx.Db.Exec("delete from surveys where id = $1 and user_id = $2", surveyId, userId)
	if err != nil {
		log.Print("Couldn't delete survey", err)
	}
}

func ReorderSurvey(surveyId int64, questionsOrder []int64, userId int64) {
	_, err := context.Ctx.Db.Exec("update surveys set updated = now(), questions_order = $1 where id = $2 and user_id = $3", pq.Array(questionsOrder), surveyId, userId)
	if err != nil {
		log.Println("ReorderSurvey", err)
	}
}

type Response map[string]string

func (a Response) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Response) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion []byte failed")
	}
	x := make(map[string]string)
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}

	*a = make(Response, len(x))
	for k, v := range x {
		(*a)[k] = v
	}

	return nil
}

func RecordResponse(surveyId int64, response Response) {
	_, err := context.Ctx.Db.Exec("insert into responses (survey_id, Response) VALUES ($1, $2)", surveyId, response)
	if err != nil {
		log.Println("RecordResponse", err)
	}
}
