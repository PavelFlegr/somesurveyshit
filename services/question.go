package services

import (
	"github.com/lib/pq"
	"log"
	"main/context"
)

func ListQuestions(surveyId int64, blockId int64) ([]Question, error) {
	rows, err := context.Ctx.Db.Query("select id, description, title, options, survey_id, block_id from questions where block_id = $1 and survey_id = $2", blockId, surveyId)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	questions := []Question{}
	for rows.Next() {
		var question Question
		err = rows.Scan(&question.Id, &question.Description, &question.Title, &question.Options, &question.SurveyId, &question.BlockId)
		if err != nil {
			log.Print(err)
			return questions, err
		}
		questions = append(questions, question)
	}

	return questions, nil
}

func ListQuestionsBySurvey(surveyId int64) ([]Question, error) {
	rows, err := context.Ctx.Db.Query("select id, description, title, options, survey_id, block_id from questions where survey_id = $1", surveyId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var questions []Question
	for rows.Next() {
		var question Question
		err = rows.Scan(&question.Id, &question.Description, &question.Title, &question.Options, &question.SurveyId, &question.BlockId)
		if err != nil {
			log.Println(err)
			return questions, err
		}
		questions = append(questions, question)
	}

	return questions, err
}

func GetQuestion(surveyId int64, questionId int64) Question {
	var question Question
	question.Id = questionId
	question.SurveyId = surveyId
	err := context.Ctx.Db.QueryRow("select description, title, options from questions where id = $1 and survey_id = $2", questionId, surveyId).Scan(&question.Description, &question.Title, &question.Options)
	if err != nil {
		log.Print(err)
	}

	return question
}

func CreateQuestion(surveyId int64, userId int64, blockId int64) Question {
	var question Question
	question.Title = "New Question"
	question.SurveyId = surveyId
	err := context.Ctx.Db.QueryRow("insert into questions (survey_id, title, user_id, block_id) values ($1, $2, $3, $4) returning id", surveyId, question.Title, userId, blockId).Scan(&question.Id)
	if err != nil {
		log.Print(err)
	}
	_, err = context.Ctx.Db.Exec("update surveys set updated = now() where id = $1", surveyId)
	if err != nil {
		log.Println(err)
	}
	_, err = context.Ctx.Db.Exec("update blocks set questions_order = array_append(questions_order, $1) where id = $2", question.Id, blockId)
	if err != nil {
		log.Println(err)
	}

	return question
}

func UpdateQuestion(surveyId int64, question Question) {
	_, err := context.Ctx.Db.Exec("update questions set title = $1, description = $2, options = $3 where id = $4 and survey_id = $5", question.Title, question.Description, question.Options, question.Id, surveyId)
	if err != nil {
		log.Println(err)
	}
	context.Ctx.Db.Exec("update surveys set updated = now() where id = $1", surveyId)
}

func DeleteQuestion(surveyId int64, questionId int64) {
	var blockId int64
	err := context.Ctx.Db.QueryRow("delete from questions where id = $1 and survey_id = $2 returning block_id", questionId, surveyId).Scan(&blockId)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = context.Ctx.Db.Exec("update surveys set updated = now() where id = $1", surveyId)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = context.Ctx.Db.Exec("update blocks set questions_order = array_remove(questions_order, $1) where id = $2", questionId, blockId)
	if err != nil {
		log.Println(err)
		return
	}
}

func ReorderQuestion(surveyId int64, questionId int64, blockId int64, index int) {
	var oldBlockId int64
	err := context.Ctx.Db.QueryRow("select block_id from questions where id = $1 and survey_id = $2", questionId, surveyId).Scan(&oldBlockId)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = context.Ctx.Db.Exec("update blocks set questions_order = array_remove(questions_order, $1) where id = $2", questionId, oldBlockId)
	if err != nil {
		log.Println(err)
		return
	}

	var questionsOrder []int64
	err = context.Ctx.Db.QueryRow("select questions_order from blocks where id = $1 and survey_id = $2", blockId, surveyId).Scan((*pq.Int64Array)(&questionsOrder))
	if err != nil {
		log.Println(err)
		return
	}

	var newOrder []int64
	if index == 0 {
		newOrder = append(newOrder, questionId)
	}
	for i := 0; i < len(questionsOrder); i++ {
		newOrder = append(newOrder, questionsOrder[i])
		if i+1 == index {
			newOrder = append(newOrder, questionId)
		}
	}

	_, err = context.Ctx.Db.Exec("update blocks set questions_order = $1 where id = $2 and survey_id = $3", pq.Array(newOrder), blockId, surveyId)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = context.Ctx.Db.Exec("update questions set block_id = $1 where id = $2 and survey_id = $3", blockId, questionId, surveyId)
	if err != nil {
		log.Println(err)
		return
	}
}
