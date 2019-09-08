package main

import (
	"encoding/json"
	"fmt"	
	"log"
	"net/http"
	"strconv"
	"strings"

	"seq/service"
)

type H map[string]interface{}

type response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data H      `json:"data"`
}

func formatJsonErr(err error) string {
	resp := response{
		Code: 1,
		Msg: err.Error(),
	}

	log.Println("error occurred: ", err)

	res, marshallErr := json.Marshal(resp)
	if marshallErr != nil {
		log.Println("marshall error: ", marshallErr)
	}

	return string(res)
}

func formatJsonSuccess(nextId int64) string {
	resp := response{
		Code: 0,
		Msg:  "ok",
		Data: H{
			"id": nextId,
		},
	}

	res, err := json.Marshal(resp)
	if err != nil {
		return formatJsonErr(err)
	}

	log.Printf("current id : %d\n", nextId)
	return string(res)
}

// 保持简单，尽量不要引入不需要的组件
func main() {
	// 返回下一个可用的id，格式为json
	http.HandleFunc("/nextId", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		currentId := service.NextId()
		log.Printf("current id : %d\n", currentId)
		fmt.Fprintf(w, formatJsonSuccess(currentId))
	
		return
	})

	// 直接返回下一个可用的id
	http.HandleFunc("/nextIdSimple", func(w http.ResponseWriter, r *http.Request) {
		currentId := service.NextId()

		log.Printf("current id : %d\n", currentId)
		fmt.Fprintf(w, "%d", currentId)

		return
	})

	var idWorkerMap = make(map[int]*service.IdWorker)

	// 采用snowflake算法生成全局唯一的id
	http.HandleFunc("/worker/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/worker/"))
		if err != nil {
			fmt.Fprintf(w, formatJsonErr(err))
			return
		}

		value, ok := idWorkerMap[id]
		if ok {
			nextId, err := value.NextId()
			if err != nil {
				fmt.Fprintf(w, formatJsonErr(err))
				return
			}

			fmt.Fprintf(w, formatJsonSuccess(nextId))
			return
		}

		iw, err := service.NewIdWorker(int64(id))
		if err != nil {
			fmt.Fprintf(w, formatJsonErr(err))
			return
		}

		nextId, err := iw.NextId()
		if err != nil {
			fmt.Fprintf(w, formatJsonErr(err))
			return
		}

		idWorkerMap[id] = iw
		fmt.Fprintf(w, formatJsonSuccess(nextId))
		return
	})

	log.Println("ID server is listening on", service.Addr())
	log.Fatal(http.ListenAndServe(service.Addr(), nil))
}