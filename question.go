package main

import (
	"github.com/lib/pq"
	"log"
)

func updateTitle(questionId, title string) {
	db.Exec("update questions set title = $1 where id = $2", title, questionId)
}

func listOptions(questionId int64) []Option {
	rows, err := db.Query("select id, question_id, value from options where question_id = $1", questionId)
	if err != nil {
		log.Print(err)
	}
	var options []Option
	for rows.Next() {
		var option Option
		rows.Scan(&option.Id, &option.QuestionId, &option.Value)
		options = append(options, option)
	}

	return options
}

func listQuestions() []Question {
	rows, err := db.Query("select id, description, title, options_order from questions")
	defer rows.Close()
	if err != nil {
		log.Print(err)
	}

	var questions []Question
	var optionsOrder []int64
	for rows.Next() {
		var question Question
		rows.Scan(&question.Id, &question.Description, &question.Title, (*pq.Int64Array)(&optionsOrder))
		tmp := make(map[int64]Option)
		question.Options = listOptions(question.Id)
		for _, v := range question.Options {
			tmp[v.Id] = v
		}
		for i, v := range optionsOrder {
			question.Options[i] = tmp[v]
		}
		questions = append(questions, question)
	}

	return questions
}

func createQuestion() Question {
	row := db.QueryRow("insert into questions default values returning id")
	var tmp Question
	err := row.Scan(&tmp.Id)
	if err != nil {
		log.Print(err)
	}

	return tmp
}

func addOption(questionId int64, value string) Option {
	row := db.QueryRow("insert into options values (DEFAULT, $1, $2) returning id", questionId, value)
	var tmp Option
	tmp.QuestionId = questionId
	tmp.Value = value
	err := row.Scan(&tmp.Id)
	if err != nil {
		log.Print(err)
	}

	db.Exec("update questions set options_order = array_append(options_order, $1) where id = $2", tmp.Id, questionId)

	return tmp
}

func deleteOption(questionId int64, optionId int64) {
	db.Exec("delete from options where id = $1 and question_id = $2", optionId, questionId)
	db.Exec("update questions set options_order = array_remove(options_order, $1) where id = $2", optionId, questionId)
}

func updateOption(questionId int64, optionId int64, value string) {
	db.Exec("update options set value = $1 where id = $2 and question_id = $3", value, optionId, questionId)
}

func updateDescription(questionId int64, description string) {
	db.Exec("update questions set description = $1 where id = $2", description, questionId)
}

func deleteQuestion(questionId int64) {
	db.Exec("delete from questions where id = $1", questionId)
}

func reorderOptions(questionId int64, oldIndex int64, newIndex int64) {
	if oldIndex == newIndex {
		return
	}
	var order []int64
	db.QueryRow("select options_order from questions where id = $1", questionId).Scan((*pq.Int64Array)(&order))
	val := order[oldIndex]
	if oldIndex > newIndex {
		for i := oldIndex - 1; i >= newIndex; i-- {
			order[i+1] = order[i]
		}
		order[newIndex] = val
	} else {
		for i := oldIndex; i < newIndex; i++ {
			order[i] = order[i+1]
		}
		order[newIndex] = val
	}

	db.Exec("update questions set options_order = $1", pq.Array(order))
}
