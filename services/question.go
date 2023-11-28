package services

import (
	"fmt"
	"log"
	"main/global"

	"github.com/lib/pq"
)

func CountQuestions(surveyId int64) int64 {
	var count int64
	err := global.Db.QueryRow("select count(1) from questions where survey_id = $1", surveyId).Scan(&count)
	if err != nil {
		log.Println(err)
	}

	return count
}

func ListQuestions(surveyId int64, blockId int64) ([]Question, error) {
	rows, err := global.Db.Query("select id, description, title, configuration, survey_id, block_id from questions where block_id = $1 and survey_id = $2", blockId, surveyId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	questions := []Question{}
	for rows.Next() {
		var question Question
		err = rows.Scan(&question.Id, &question.Description, &question.Title, &question.Configuration, &question.SurveyId, &question.BlockId)
		if err != nil {
			log.Print(err)
			return questions, err
		}
		questions = append(questions, question)
	}

	return questions, nil
}

func ListQuestionsBySurvey(surveyId int64) ([]Question, error) {
	rows, err := global.Db.Query("select id, description, title, configuration, survey_id, block_id from questions where survey_id = $1", surveyId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var questions []Question
	for rows.Next() {
		var question Question
		err = rows.Scan(&question.Id, &question.Description, &question.Title, &question.Configuration, &question.SurveyId, &question.BlockId)
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
	err := global.Db.QueryRow("select description, title, configuration, block_id from questions where id = $1 and survey_id = $2", questionId, surveyId).Scan(&question.Description, &question.Title, &question.Configuration, &question.BlockId)
	if err != nil {
		log.Print(err)
	}

	return question
}

func CreateQuestion(surveyId int64, userId int64, blockId int64, configuration Configuration) Question {
	questionCount := CountQuestions(surveyId)
	var question Question
	question.Title = fmt.Sprintf("Question %v", questionCount+1)
	question.SurveyId = surveyId
	question.BlockId = blockId
	question.Configuration = configuration
	err := global.Db.QueryRow("insert into questions (survey_id, title, user_id, block_id, configuration) values ($1, $2, $3, $4, $5) returning id", surveyId, question.Title, userId, blockId, configuration).Scan(&question.Id)
	if err != nil {
		log.Print(err)
	}
	_, err = global.Db.Exec("update surveys set updated = now() where id = $1", surveyId)
	if err != nil {
		log.Println(err)
	}
	_, err = global.Db.Exec("update blocks set questions_order = array_append(questions_order, $1) where id = $2", question.Id, blockId)
	if err != nil {
		log.Println(err)
	}

	return question
}

func UpdateQuestion(surveyId int64, question *Question) {
	err := global.Db.QueryRow("update questions set title = $1, description = $2, configuration = $3 where id = $4 and survey_id = $5 returning block_id", question.Title, question.Description, question.Configuration, question.Id, surveyId).Scan(&question.BlockId)
	if err != nil {
		log.Println(err)
	}
	global.Db.Exec("update surveys set updated = now() where id = $1", surveyId)
}

func DeleteQuestion(surveyId int64, questionId int64) {
	var blockId int64
	err := global.Db.QueryRow("delete from questions where id = $1 and survey_id = $2 returning block_id", questionId, surveyId).Scan(&blockId)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = global.Db.Exec("update surveys set updated = now() where id = $1", surveyId)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = global.Db.Exec("update blocks set questions_order = array_remove(questions_order, $1) where id = $2", questionId, blockId)
	if err != nil {
		log.Println(err)
		return
	}
}

func ReorderQuestion(surveyId int64, questionId int64, blockId int64, index int) {
	var oldBlockId int64
	err := global.Db.QueryRow("select block_id from questions where id = $1 and survey_id = $2", questionId, surveyId).Scan(&oldBlockId)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = global.Db.Exec("update blocks set questions_order = array_remove(questions_order, $1) where id = $2", questionId, oldBlockId)
	if err != nil {
		log.Println(err)
		return
	}

	var questionsOrder []int64
	err = global.Db.QueryRow("select questions_order from blocks where id = $1 and survey_id = $2", blockId, surveyId).Scan((*pq.Int64Array)(&questionsOrder))
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

	_, err = global.Db.Exec("update blocks set questions_order = $1 where id = $2 and survey_id = $3", pq.Array(newOrder), blockId, surveyId)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = global.Db.Exec("update questions set block_id = $1 where id = $2 and survey_id = $3", blockId, questionId, surveyId)
	if err != nil {
		log.Println(err)
		return
	}
}
