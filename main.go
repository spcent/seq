package main

import (
	"encoding/json"
	"fmt"	
	"log"
	"net/http"

	"seq/service"
)

type H map[string]interface{}

type response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data H      `json:"data"`
}

func main() {
	http.HandleFunc("/nextId", nextId)
	http.HandleFunc("/nextIdSimple", nextIdSimple)

	log.Println("ID server is listening on", service.Addr())
	log.Fatal(http.ListenAndServe(service.Addr(), nil))
}

func nextIdSimple(w http.ResponseWriter, r *http.Request) {
	currentId := service.NextId()

	log.Printf("current id : %d\n", currentId)
	fmt.Fprintf(w, "%d", currentId)

	return
}

func nextId(w http.ResponseWriter, r *http.Request) {
	currentId := service.NextId()

	w.Header().Set("Content-Type", "application/json")
	log.Printf("current id : %d\n", currentId)
	resp := response{
		Code: 0,
		Msg:  "ok",
		Data: H{
			"id": currentId,
		},
	}

	res, _ := json.Marshal(resp)
	fmt.Fprintf(w, string(res))

	return
}