package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile | log.LUTC | log.Lmsgprefix)
	log.SetPrefix("[webapp] ")
	log.Println("Logger initialized")
}

func main() {
	http.HandleFunc("/", processInferenceRequest)
	http.HandleFunc("/management/throughputUnit", processManagementRequest)
	http.ListenAndServe(":8080", nil)
}

func processInferenceRequest(w http.ResponseWriter, req *http.Request) {
	log.Println("Inference request received")

	modelName := req.URL.Query().Get("modelName")
	if len(modelName) == 0 {
		log.Println("No modelName query parameter")
		http.Error(w, "No modelName query parameter", http.StatusBadRequest)
		return
	}

	reqUrl := &url.URL{}
	reqUrl.Scheme = "http"
	targetHost := getEnv("TARGET_HOST", "localhost")
	targetPort := getEnv("TARGET_PORT", "8081")
	reqUrl.Host = fmt.Sprintf("%s:%s", targetHost, targetPort)

	// Relay modelName information to compute
	// This information will be used by Istio routing rule
	q := url.Values{}
	q.Add("modelName", modelName)
	reqUrl.RawQuery = q.Encode()

	rsp, err := http.Get(reqUrl.String())
	if err != nil {
		log.Println("Request sent to compute failed")
		log.Println(err)
		return
	}

	defer rsp.Body.Close()

	log.Println("Request sent to compute succeeded")
	body, _ := ioutil.ReadAll(rsp.Body)
	fmt.Fprintln(w, string(body))
}

func processManagementRequest(w http.ResponseWriter, req *http.Request) {
	log.Println("Management request received")

	switch req.Method {
	case "GET":
		modelName := req.URL.Query().Get("modelName")
		if len(modelName) == 0 {
			rsp, err := getAllEntities()
			if err != nil {
				log.Println("getAllEntities() failed")
				http.Error(w, "Failed to get entities", http.StatusInternalServerError)
				return
			}
			fmt.Fprintln(w, rsp)
			return
		}

		rsp, err := getEntity(modelName)
		if err != nil {
			log.Println("getAllEntity() failed")
			errMsg := fmt.Sprintf("Failed to get entity for %s", modelName)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, rsp)
		return

	case "PATCH":
		modelName := req.URL.Query().Get("modelName")
		if len(modelName) == 0 {
			log.Println("No modelName query parameter")
			http.Error(w, "No modelName query parameter", http.StatusBadRequest)
			return
		}

		tu := req.URL.Query().Get("throughputUnit")
		if len(tu) == 0 {
			log.Println("No throughputUnit query parameter")
			http.Error(w, "No throughputUnit query parameter", http.StatusBadRequest)
			return
		}
		throughputUnit, err := strconv.Atoi(tu)
		if err != nil {
			log.Println("ThroughputUnit is not integer")
			http.Error(w, "ThroughputUnit is not integer", http.StatusBadRequest)
			return
		}

		err = updateEntity(modelName, throughputUnit)
		if err != nil {
			log.Println("updateEntity() failed")
			errMsg := fmt.Sprintf("Failed to update entity for %s", modelName)
			http.Error(w, errMsg, http.StatusInternalServerError)
			return
		}
		return

	default:
		log.Println("Invalid http method")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func getEntity(modelName string) (string, error) {
	return "entity", nil
}

func getAllEntities() (string, error) {
	return "all entities", nil
}

func updateEntity(modelName string, throughputUnit int) error {
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
