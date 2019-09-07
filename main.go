package main

import (
	"encoding/json"
	"fmt"	
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"

	"github.com/satori/go.uuid"

	"seq/service"
)

var (
	startID int64
	curID   int64
	maxID   int64
	uid     = uuid.Must(uuid.NewV4(), nil)
	host, _ = os.Hostname()
	conf    *service.Config
	mu      sync.Mutex
)

type H map[string]interface{}

type response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data H      `json:"data"`
}

func init() {
	conf = service.NewConfig()
	service.InitDB(conf.MySQL)
}

func main() {
	http.HandleFunc("/nextId", nextId)
	http.HandleFunc("/nextIdSimple", nextIdSimple)

	log.Println("ID server is listening on", conf.PORT)
	log.Fatal(http.ListenAndServe(conf.PORT, nil))
}

func genId() int64 {
	s := atomic.LoadInt64(&startID)
	if s == 0 || curID == maxID {
		mu.Lock()
		s = atomic.LoadInt64(&startID)
		if s == 0 || curID == maxID {
			id, err := service.New(uid)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("get new id : %d for ip : %s，uuid : %s\n", id, host, uid)

			atomic.StoreInt64(&startID, id*conf.STEP)
			atomic.StoreInt64(&curID, id*conf.STEP)
			atomic.StoreInt64(&maxID, (id+1)*conf.STEP)
		}

		mu.Unlock()
		log.Println("start id get =====>", startID)
	}

	// 有点暴力，需要压力测试看看效果
	for {
		if atomic.CompareAndSwapInt64(&curID, curID, curID+1) {
			// quit only when CompareAndSwap success, otherwise retry
			return curID
		}

		log.Printf("incr current id: %d failed", curID)
	}

	return -1
}

func nextIdSimple(w http.ResponseWriter, r *http.Request) {
	currentId := genId()

	log.Printf("current id : %d\n", currentId)
	fmt.Fprintf(w, "%d", currentId)

	return
}

func nextId(w http.ResponseWriter, r *http.Request) {
	currentId := genId()

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