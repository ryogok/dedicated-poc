package model

import (
	"log"

	"github.com/ryogok/dedicated-poc/pkg/k8sapi"
	"github.com/ryogok/dedicated-poc/pkg/table"
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

	pinfo, err := table.UpdateEntity(modelName, throughputUnit)
	if err != nil {
		return err
	}

	return k8sapi.UpdateDeployment(modelName, pinfo)
}
