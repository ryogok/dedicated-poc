package main

import (
	"fmt"
	"log"
	"net/http"
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile | log.LUTC | log.Lmsgprefix)
	log.SetPrefix("[compute]")
	log.Println("logger initialized")
}

func main() {
	http.HandleFunc("/", processGet)
	http.ListenAndServe(":8081", nil)
}

func processGet(w http.ResponseWriter, req *http.Request) {
	log.Println("request received")
	fmt.Fprintln(w, "I'm compute")
}
