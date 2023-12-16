package services

import (
	"fmt"
	"log"
	"main/global"
	"strings"

	"github.com/lib/pq"
)

func CountBlocks(surveyId int64) int64 {
	var count int64
	err := global.Db.QueryRow("select count(1) from blocks where survey_id = $1", surveyId).Scan(&count)
	if err != nil {
		log.Println(err)
	}
	return count
}

func CreateBlock(block *Block, userId int64) error {
	var err = global.Db.QueryRow("insert into blocks (user_id, survey_id, title, created) values ($1, $2, $3, now()) returning id, submit_after", userId, block.SurveyId, block.Title).Scan(&block.Id, &block.SubmitAfter)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = global.Db.Exec("update surveys set updated = now(), blocks_order = array_append(blocks_order, $1) where id = $2", block.Id, block.SurveyId)
	if err != nil {
		log.Println(err)
	}

	return err
}

func ListBlocks(surveyId int64) ([]Block, error) {
	var rows, err = global.Db.Query("select id, title, randomize, submit, submit_after from blocks where survey_id = $1", surveyId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var blocks []Block
	for rows.Next() {
		var block = Block{SurveyId: surveyId}
		err = rows.Scan(&block.Id, &block.Title, &block.Randomize, &block.Submit, &block.SubmitAfter)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		blocks = append(blocks, block)
	}

	return blocks, err
}

func GetBlock(blockId int64, surveyId int64) (Block, error) {
	block := Block{Id: blockId}
	var questionsOrder []int64
	var err = global.Db.QueryRow("select id, survey_id, title, randomize, submit, submit_after, questions_order from blocks where id = $1 and survey_id = $2", blockId, surveyId).Scan(
		&block.Id,
		&block.SurveyId,
		&block.Title,
		&block.Randomize,
		&block.Submit,
		&block.SubmitAfter,
		(*pq.Int64Array)(&questionsOrder),
	)
	if err != nil {
		log.Println(err)
		return block, err
	}

	tmp := make(map[int64]Question)
	block.Questions, err = ListQuestions(surveyId, blockId)
	if err != nil {
		log.Println(err)
		return block, err
	}
	for _, v := range block.Questions {
		tmp[v.Id] = v
	}
	for i, v := range questionsOrder {
		block.Questions[i] = tmp[v]
	}

	return block, err
}

func RemoveBlock(surveyId int64, blockId int64) error {
	var _, err = global.Db.Exec("delete from blocks where id = $1 and survey_id = $2", blockId, surveyId)
	if err != nil {
		log.Println(err)
	}

	_, err = global.Db.Exec("update surveys set updated = now(), blocks_order = array_remove(blocks_order, $1) where id = $2", blockId, surveyId)
	if err != nil {
		log.Println(err)
	}

	_, err = global.Db.Exec("update surveys set blocks_order = array_remove(blocks_order, $1) where id = $2", blockId, surveyId)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

func RenameBlock(blockId int64, surveyId int64, title string) error {
	_, err := global.Db.Exec("update blocks set title = $1 where id = $2 and survey_id = $3", title, blockId, surveyId)

	return err
}

type Record struct {
	Key   string
	Value any
}

func UpdateBlock(blockId int64, surveyId int64, data []Record) error {
	if len(data) == 0 {
		return nil
	}

	fields := []string{}
	values := []any{blockId, surveyId}
	for i := range data {
		fields = append(fields, fmt.Sprintf("%v = $%v", data[i].Key, i+3))
		values = append(values, data[i].Value)
	}
	query := fmt.Sprintf("update blocks set %v where id = $1 and survey_id = $2", strings.Join(fields, ","))
	_, err := global.Db.Exec(query, values...)

	return err
}

func SetRandomize(blockId int64, surveyId int64, randomize bool) error {
	_, err := global.Db.Exec("update blocks set randomize = $1 where id = $2 and survey_id = $3", randomize, blockId, surveyId)

	return err
}
