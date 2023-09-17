package services

import (
	"log"
	"main/context"
)

func ListQuestions(surveyId int64) []Question {
	rows, err := context.Ctx.Db.Query("select id, description, title, options, survey_id from questions where survey_id = $1", surveyId)
	defer rows.Close()
	if err != nil {
		log.Print("ListQuestions", err)
	}

	var questions []Question
	for rows.Next() {
		var question Question
		rows.Scan(&question.Id, &question.Description, &question.Title, &question.Options, &question.SurveyId)
		questions = append(questions, question)
	}

	return questions
}

func GetQuestion(surveyId int64, questionId int64) Question {
	var question Question
	question.Id = questionId
	question.SurveyId = surveyId
	err := context.Ctx.Db.QueryRow("select description, title, options from questions where id = $1 and survey_id = $2", questionId, surveyId).Scan(&question.Description, &question.Title, &question.Options)
	if err != nil {
		log.Print("GetQuestion", err)
	}

	return question
}

func CreateQuestion(surveyId int64, userId int64) Question {
	var question Question
	question.Title = "New Question"
	question.SurveyId = surveyId
	row := context.Ctx.Db.QueryRow("insert into questions (survey_id, title, user_id) values ($1, $2, $3) returning id", surveyId, question.Title, userId)
	err := row.Scan(&question.Id)
	if err != nil {
		log.Print("CreateQuestion", err)
	}
	context.Ctx.Db.Exec("update surveys set updated = now(), questions_order = array_append(questions_order, $1) where id = $2", question.Id, surveyId)

	return question
}

func UpdateQuestion(surveyId int64, question Question) {
	_, err := context.Ctx.Db.Exec("update questions set title = $1, description = $2, options = $3 where id = $4 and survey_id = $5", question.Title, question.Description, question.Options, question.Id, surveyId)
	if err != nil {
		log.Println("UpdateQuestion", err)
	}
	context.Ctx.Db.Exec("update surveys set updated = now() where id = $1", surveyId)
}

func DeleteQuestion(surveyId int64, questionId int64) {
	_, err := context.Ctx.Db.Exec("delete from questions where id = $1 and survey_id = $2", questionId, surveyId)
	if err != nil {
		log.Println("DeleteQuestion", err)
	}
	context.Ctx.Db.Exec("update surveys set updated = now(), questions_order = array_remove(questions_order, $1) where id = $2", questionId, surveyId)
}
