package table

import (
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
)

var client *aztables.Client

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
	_, err = serviceClient.CreateTable(context.TODO(), tableName, nil)
	if err != nil {
		log.Println("Failed to create a table")
		panic(err)
	}

	client = serviceClient.NewClient(tableName)
}

func GetEntity(modelName string) (string, error) {
	return "entity", nil
}

func GetAllEntities() (string, error) {
	return "all entities", nil
}

func UpdateEntity(modelName string, throughputUnit int) error {
	return nil
}
