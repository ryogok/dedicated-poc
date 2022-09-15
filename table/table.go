// Reference: https://docs.microsoft.com/en-us/azure/cosmos-db/table/how-to-use-go?tabs=bash

package table

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
	"webapp/types"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/google/uuid"
)

var modelClient *aztables.Client
var partitionClient *aztables.Client
var partitionKey string

type ModelEntity struct {
	aztables.Entity
	ModelName             string
	PartitionName         string
	RequestedTU           int32
	RequestedTULastUpdate aztables.EDMDateTime
}

type PartitionEntity struct {
	aztables.Entity
	PartitionName string
	ModelNames    string
	TUCurrent     int32
	TUCapacity    int32
	TimeStamp     aztables.EDMDateTime
}

func init() {
	// Initialize logger
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile | log.LUTC | log.Lmsgprefix)
	log.SetPrefix("[table] ")
	log.Println("Logger initialized")

	connStr, ok := os.LookupEnv("AZURE_STORAGE_CONNECTION_STRING")
	if !ok {
		log.Println("AZURE_STORAGE_CONNECTION_STRING environment variable not found")
		panic("AZURE_STORAGE_CONNECTION_STRING environment variable not found")
	}

	serviceClient, err := aztables.NewServiceClientFromConnectionString(connStr, nil)
	if err != nil {
		log.Println("Failed to create a service client object")
		fmt.Println(err)
		panic(err)
	}

	// Create tables
	modelTableName := "Model"
	/*
		_, err = serviceClient.CreateTable(context.TODO(), modelTableName, nil)
		if err != nil {
			log.Println("Failed to create a table")
			fmt.Println(err)
			panic(err)
		}
	*/
	modelClient = serviceClient.NewClient(modelTableName)

	partitionTableName := "Partition"
	/*
		_, err = serviceClient.CreateTable(context.TODO(), partitionTableName, nil)
		if err != nil {
			log.Println("Failed to create a table")
			fmt.Println(err)
			panic(err)
		}
	*/
	partitionClient = serviceClient.NewClient(partitionTableName)

	partitionKey = "partitionkey"
}

func GetEntity(modelName string) (string, error) {
	log.Println("GetEntity() called")

	return "entity", nil
}

func GetAllEntities() (string, error) {
	log.Println("GetAllEntities() called")

	filter := fmt.Sprintf("PartitionKey eq '%s'", partitionKey)
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
		Select: to.Ptr("ModelName,RequestedTU,RequestedTULastUpdate"),
		Top:    to.Ptr(int32(10)),
	}

	pager := modelClient.NewListEntitiesPager(options)
	for pager.More() {
		rsp, err := pager.NextPage(context.Background())
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		for _, e := range rsp.Entities {
			var entity ModelEntity
			err = json.Unmarshal(e, &entity)
			if err != nil {
				fmt.Println(err)
				return "", err
			}

			log.Printf("ModelName:%s, RequestedTU:%v", entity.ModelName, entity.RequestedTU)
		}
	}

	// TODO: return json string
	return "all entities", nil
}

func UpdateEntity(modelName string, throughputUnit int) (*types.PartitionInfo, error) {
	log.Println("UpdateEntity() called")

	isNew, err := isNewModel(modelName)
	if err != nil {
		log.Println("isNewModel() failed")
		return nil, err
	}

	if isNew {
		log.Printf("No existing entity for modelName:%s. Adding a new one...", modelName)

		err = addNewModelEntity(modelName, throughputUnit)
		if err != nil {
			log.Println("Failed to add a new model entity")
			return nil, err
		}
		log.Println("Added a new model entity successfully")

		pinfo, err := addNewModelToPartitionEntity(modelName, throughputUnit)
		if err != nil {
			log.Println("Failed to add the new model to partition entity")
			return nil, err
		}
		log.Println("Added a new model to partition entity successfully")

		return pinfo, nil
	}

	modelEntity, err := getModelEntity(modelName)
	if err != nil {
		log.Println("getModelEntity() failed")
		return nil, err
	}

	if throughputUnit != int(modelEntity.RequestedTU) {
		log.Printf("New TU requested for modelName:%s. Updating...", modelName)

		err = upsertModelEntity(modelEntity.RowKey, modelName, throughputUnit)
		if err != nil {
			log.Println("Failed to upsert a model entity")
			return nil, err
		}

		pinfo := &types.PartitionInfo{
			Name:  "TODO",
			IsNew: false,
		}

		log.Println("Upserted an entry successfully")
		return pinfo, nil
	}

	pinfo := &types.PartitionInfo{
		Name:  "TODO",
		IsNew: false,
	}

	return pinfo, nil
}

func isNewModel(modelName string) (bool, error) {
	filter := fmt.Sprintf("PartitionKey eq '%s' and ModelName eq '%s'", partitionKey, modelName)
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
		Select: to.Ptr("RowKey"),
		Top:    to.Ptr(int32(10)),
	}

	pager := modelClient.NewListEntitiesPager(options)
	rsp, err := pager.NextPage(context.Background())
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	return len(rsp.Entities) == 0, nil
}

func getModelEntity(modelName string) (*ModelEntity, error) {
	filter := fmt.Sprintf("PartitionKey eq '%s' and ModelName eq '%s'", partitionKey, modelName)
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
		Select: to.Ptr("RowKey,ModelName,PartitionName,RequestedTU,RequestedTULastUpdate"),
		Top:    to.Ptr(int32(10)),
	}

	pager := modelClient.NewListEntitiesPager(options)
	rsp, err := pager.NextPage(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if len(rsp.Entities) == 0 {
		return nil, errors.New("No existing entry for model:" + modelName)
	}

	var entity ModelEntity
	e := rsp.Entities[0]
	err = json.Unmarshal(e, &entity)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &entity, nil
}

func addNewModelEntity(modelName string, throughputUnit int) error {
	entity := ModelEntity{
		Entity: aztables.Entity{
			PartitionKey: partitionKey,
			RowKey:       uuid.New().String(),
		},
		ModelName:             modelName,
		PartitionName:         "", // TODO
		RequestedTU:           int32(throughputUnit),
		RequestedTULastUpdate: aztables.EDMDateTime(time.Now()),
	}

	marshalled, err := json.Marshal(entity)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = modelClient.AddEntity(context.TODO(), marshalled, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func upsertModelEntity(rowKey string, modelName string, throughputUnit int) error {
	entity := ModelEntity{
		Entity: aztables.Entity{
			PartitionKey: partitionKey,
			RowKey:       rowKey,
		},
		ModelName:             modelName,
		PartitionName:         "", // TODO
		RequestedTU:           int32(throughputUnit),
		RequestedTULastUpdate: aztables.EDMDateTime(time.Now()),
	}

	marshalled, err := json.Marshal(entity)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = modelClient.UpsertEntity(context.TODO(), marshalled, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func addNewModelToPartitionEntity(modelName string, throughputUnit int) (*types.PartitionInfo, error) {
	partitionEntity, err := tryFindVacantPartition(throughputUnit)
	if err != nil {
		log.Println("tryFindVacantPartition() failed")
		return nil, err
	}

	if partitionEntity == nil {
		partitionName, err := addNewPartitionEntity(modelName, throughputUnit)
		if err != nil {
			log.Println("addNewPartitionEntity() failed")
			return nil, err
		}

		pinfo := &types.PartitionInfo{
			Name:  partitionName,
			IsNew: true,
		}

		return pinfo, nil
	}

	err = upsertPartitionEntity(partitionEntity, modelName, throughputUnit)
	if err != nil {
		log.Println("upsertPartitionEntity() failed")
		return nil, err
	}

	pinfo := &types.PartitionInfo{
		Name:  partitionEntity.PartitionName,
		IsNew: false,
	}

	return pinfo, nil
}

func tryFindVacantPartition(throughputUnit int) (*PartitionEntity, error) {
	filter := fmt.Sprintf("PartitionKey eq '%s'", partitionKey)
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
		Select: to.Ptr("RowKey,PartitionName,ModelNames,TUCurrent,TUCapacity"),
		Top:    to.Ptr(int32(10)),
	}

	pager := partitionClient.NewListEntitiesPager(options)
	rsp, err := pager.NextPage(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for _, e := range rsp.Entities {
		var entity PartitionEntity
		err = json.Unmarshal(e, &entity)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if entity.TUCapacity-entity.TUCurrent >= int32(throughputUnit) {
			return &entity, nil
		}
	}

	return nil, nil
}

func addNewPartitionEntity(modelName string, throughputUnit int) (string, error) {
	// For demo: we want to assign "p1" or "p2" only for the partition name
	partitionName := "p1"
	exists, err := partitionEntityExists("p1")
	if err != nil {
		log.Println("partitionEntityExists() failed")
		return "", err
	}
	if exists {
		partitionName = "p2"
	}

	entity := PartitionEntity{
		Entity: aztables.Entity{
			PartitionKey: partitionKey,
			RowKey:       uuid.New().String(),
		},
		PartitionName: partitionName,
		ModelNames:    modelName,
		TUCurrent:     int32(throughputUnit),
		TUCapacity:    int32(getNecessaryTUCapacity(throughputUnit)),
		TimeStamp:     aztables.EDMDateTime(time.Now()),
	}

	marshalled, err := json.Marshal(entity)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	_, err = partitionClient.AddEntity(context.TODO(), marshalled, nil)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return partitionName, nil

}

func getNecessaryTUCapacity(throughputUnit int) int {
	TU_PER_POD := 10

	if throughputUnit%TU_PER_POD == 0 {
		return throughputUnit
	}

	return throughputUnit + TU_PER_POD - (throughputUnit % TU_PER_POD)
}

// Just for demo purpose
func partitionEntityExists(partitionName string) (bool, error) {
	filter := fmt.Sprintf("PartitionKey eq '%s' and PartitionName eq '%s'", partitionKey, partitionName)
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
		Select: to.Ptr("RowKey"),
		Top:    to.Ptr(int32(10)),
	}

	pager := partitionClient.NewListEntitiesPager(options)
	rsp, err := pager.NextPage(context.Background())
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	return len(rsp.Entities) != 0, nil
}

func upsertPartitionEntity(partitionEntity *PartitionEntity, modelName string, throughputUnit int) error {
	entity := PartitionEntity{
		Entity: aztables.Entity{
			PartitionKey: partitionKey,
			RowKey:       partitionEntity.RowKey,
		},
		PartitionName: partitionEntity.PartitionName,
		ModelNames:    partitionEntity.ModelNames + "," + modelName,
		TUCurrent:     partitionEntity.TUCurrent + int32(throughputUnit),
		TUCapacity:    int32(getNecessaryTUCapacity(int(partitionEntity.TUCurrent) + throughputUnit)),
		TimeStamp:     aztables.EDMDateTime(time.Now()),
	}

	marshalled, err := json.Marshal(entity)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = partitionClient.UpsertEntity(context.TODO(), marshalled, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
