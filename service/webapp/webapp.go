package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile | log.LUTC | log.Lmsgprefix)
	log.SetPrefix("[webapp]")
	log.Println("logger initialized")
}

func main() {
	http.HandleFunc("/", hello)
	http.ListenAndServe(":8080", nil)
}

func hello(w http.ResponseWriter, req *http.Request) {
	log.Println("request received")

	url := &url.URL{}
	url.Scheme = "http"
	targetHost := getEnv("TARGET_HOST", "localhost")
	targetPort := getEnv("TARGET_PORT", "8081")
	url.Host = fmt.Sprintf("%s:%s", targetHost, targetPort)
	urlStr := url.String()

	rsp, err := http.Get(urlStr)
	if err != nil {
		log.Println("request sent to compute failed")
		log.Println(err)
		return
	}

	defer rsp.Body.Close()

	log.Println("request sent to compute succeeded")
	body, _ := ioutil.ReadAll(rsp.Body)
	fmt.Fprintln(w, string(body))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
