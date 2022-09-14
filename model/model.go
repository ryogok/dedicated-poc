package model

import (
	"log"
	"webapp/table"
)

func GetTUForModel(modelName string) (string, error) {
	log.Println("GetTUForModel() called")
	return table.GetEntity(modelName)
}

func GetAllTUs() (string, error) {
	log.Println("GetAllTUs() called")
	return table.GetAllEntities()
}

func UpdateTUForModel(modelName string, throughputUnit int) error {
	log.Println("UpdateTUForModel() called")
	return table.UpdateEntity(modelName, throughputUnit)
}
