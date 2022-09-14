package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile | log.LUTC | log.Lmsgprefix)
	log.SetPrefix("[compute] ")
	log.Println("Logger initialized")
}

func main() {
	http.HandleFunc("/", processGet)
	http.ListenAndServe(":8081", nil)
}

func processGet(w http.ResponseWriter, req *http.Request) {
	log.Println("Request received")

	modelName := req.URL.Query().Get("modelName")
	if len(modelName) == 0 {
		log.Println("No modelName query parameter")
		return
	}

	podName := getEnv("POD_NAME", "unknown")

	res := fmt.Sprintf("Compute %s processed a request for modelName %s", podName, modelName)
	fmt.Fprintln(w, res)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
