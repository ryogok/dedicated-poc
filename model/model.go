package model

import (
	"webapp/table"
)

func GetTUForModel(modelName string) (string, error) {
	return table.GetEntity(modelName)
}

func GetAllTUs() (string, error) {
	return table.GetAllEntities()
}

func UpdateTUForModel(modelName string, throughputUnit int) error {
	return table.UpdateEntity(modelName, throughputUnit)
}
