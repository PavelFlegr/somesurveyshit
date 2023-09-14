package services

import (
	"log"
	"main/context"
)

func ListQuestions(surveyId int64) []Question {
	rows, err := context.Ctx.Db.Query("select id, description, title, options from questions where survey_id = $1", surveyId)
	defer rows.Close()
	if err != nil {
		log.Print(err)
	}

	var questions []Question
	for rows.Next() {
		var question Question
		rows.Scan(&question.Id, &question.Description, &question.Title, &question.Options)
		questions = append(questions, question)
	}

	return questions
}

func GetQuestion(questionId int64) Question {
	var question Question
	question.Id = questionId
	err := context.Ctx.Db.QueryRow("select description, title, options from questions where id = $1", questionId).Scan(&question.Description, &question.Title, &question.Options)
	if err != nil {
		log.Print(err)
	}

	return question
}

func CreateQuestion(surveyId int64) Question {
	var question Question
	question.Title = "New Question"
	row := context.Ctx.Db.QueryRow("insert into questions (survey_id, title) values ($1, $2) returning id", surveyId, question.Title)
	err := row.Scan(&question.Id)
	if err != nil {
		log.Print(err)
	}
	context.Ctx.Db.Exec("update surveys set updated = now(), questions_order = array_append(questions_order, $1) where id = $2", question.Id, surveyId)

	return question
}

func UpdateQuestion(question Question) {
	var surveyId int64
	err := context.Ctx.Db.QueryRow("update questions set title = $1, description = $2, options = $3 where id = $4 returning survey_id", question.Title, question.Description, question.Options, question.Id).Scan(&surveyId)
	if err != nil {
		log.Println(err)
	}
	context.Ctx.Db.Exec("update surveys set updated = now() where id = $1", surveyId)
}

func DeleteQuestion(questionId int64) {
	var surveyId int64
	context.Ctx.Db.QueryRow("delete from questions where id = $1 returning survey_id", questionId).Scan(&surveyId)
	context.Ctx.Db.Exec("update surveys set updated = now(), questions_order = array_remove(questions_order, $1) where id = $2", questionId, surveyId)
}
