package table

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/google/uuid"
)

var client *aztables.Client
var partitionKey string

type ThroughputUnitStateEntity struct {
	aztables.Entity
	ModelName             string
	RequestedTU           int32
	RequestedTULastUpdate aztables.EDMDateTime
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
		panic(err)
	}

	// Create table - do nothing if the table already exists
	tableName := "ThroughputUnitState"
	/*
		_, err = serviceClient.CreateTable(context.TODO(), tableName, nil)
		if err != nil {
			log.Println("Failed to create a table")
			panic(err)
		}
	*/

	client = serviceClient.NewClient(tableName)
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

	pager := client.NewListEntitiesPager(options)
	for pager.More() {
		rsp, err := pager.NextPage(context.Background())
		if err != nil {
			log.Println("page.NextPage() failed")
			return "", err
		}

		for _, e := range rsp.Entities {
			var entity ThroughputUnitStateEntity
			err = json.Unmarshal(e, &entity)
			if err != nil {
				log.Println("json.Ummarshal() failed")
				return "", err
			}

			log.Printf("ModelName:%s, RequestedTU:%v", entity.ModelName, entity.RequestedTU)
		}
	}

	// TODO: return json string
	return "all entities", nil
}

func UpdateEntity(modelName string, throughputUnit int) error {
	log.Println("UpdateEntity() called")

	filter := fmt.Sprintf("PartitionKey eq '%s' and ModelName eq '%s'", partitionKey, modelName)
	options := &aztables.ListEntitiesOptions{
		Filter: &filter,
		Select: to.Ptr("RowKey,ModelName,RequestedTU,RequestedTULastUpdate"),
		Top:    to.Ptr(int32(10)),
	}

	pager := client.NewListEntitiesPager(options)
	rsp, err := pager.NextPage(context.Background())
	if err != nil {
		log.Println("page.NextPage() failed")
		return err
	}

	if len(rsp.Entities) == 0 {
		log.Printf("No existing entity for modelName:%s. Adding a new one...", modelName)

		err = addNewEntity(modelName, throughputUnit)
		if err != nil {
			log.Println("Failed to add a new entity")
			return err
		}

		log.Println("Added a new entry successfully")
		return nil
	}

	for _, e := range rsp.Entities {
		var entity ThroughputUnitStateEntity
		err = json.Unmarshal(e, &entity)
		if err != nil {
			log.Println("json.Ummarshal() failed")
			return err
		}

		if throughputUnit != int(entity.RequestedTU) {
			log.Printf("New TU requested for modelName:%s. Updating...", modelName)

			err = upsertEntity(entity.RowKey, modelName, throughputUnit)
			if err != nil {
				log.Println("Failed to upsert an entity")
				return err
			}
		}

		log.Println("Upserted an entry successfully")
		break // we're interested in the first query result only
	}

	return nil
}

func addNewEntity(modelName string, throughputUnit int) error {
	entity := ThroughputUnitStateEntity{
		Entity: aztables.Entity{
			PartitionKey: partitionKey,
			RowKey:       uuid.New().String(),
		},
		ModelName:             modelName,
		RequestedTU:           int32(throughputUnit),
		RequestedTULastUpdate: aztables.EDMDateTime(time.Now()),
	}

	marshalled, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	_, err = client.AddEntity(context.TODO(), marshalled, nil)
	if err != nil {
		return err
	}

	return nil
}

func upsertEntity(rowKey string, modelName string, throughputUnit int) error {
	entity := ThroughputUnitStateEntity{
		Entity: aztables.Entity{
			PartitionKey: partitionKey,
			RowKey:       rowKey,
		},
		ModelName:             modelName,
		RequestedTU:           int32(throughputUnit),
		RequestedTULastUpdate: aztables.EDMDateTime(time.Now()),
	}

	marshalled, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	_, err = client.UpsertEntity(context.TODO(), marshalled, nil)
	if err != nil {
		return err
	}

	return nil
}
