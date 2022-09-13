package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
	url.Host = "localhost:8081"
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
