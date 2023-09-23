package services

import (
	"github.com/lib/pq"
	"log"
	"main/global"
)

func CreateBlock(block *Block, userId int64) error {
	var err = global.Db.QueryRow("insert into blocks (user_id, survey_id, title, created) values ($1, $2, $3, now()) returning id", userId, block.SurveyId, block.Title).Scan(&block.Id)
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
	var rows, err = global.Db.Query("select id, title from blocks where survey_id = $1", surveyId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var blocks []Block
	for rows.Next() {
		var block = Block{SurveyId: surveyId}
		err = rows.Scan(&block.Id, &block.Title)
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
	var err = global.Db.QueryRow("select id, survey_id, title, questions_order from blocks where id = $1 and survey_id = $2", blockId, surveyId).Scan(&block.Id, &block.SurveyId, &block.Title, (*pq.Int64Array)(&questionsOrder))
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
	if err != nil {
		log.Println(err)
	}

	return err
}
