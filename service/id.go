// package service
//
// https://mp.weixin.qq.com/s/F7WTNeC3OUr76sZARtqRjw

package service

import (
	"log"
	"os"
	"sync"
	"sync/atomic"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/satori/go.uuid"
)

var (
	err error
	db  *sql.DB

	startID int64
	curID   int64
	maxID   int64
	uid     = uuid.Must(uuid.NewV4(), nil)
	host, _ = os.Hostname()
	conf    *Config
	mu      sync.Mutex
)

func init() {
	conf = NewConfig()
	initDB(conf.MySQL)
}

func initDB(conf MySQL) {
	dsn := conf.User + ":" + conf.PassWord + "@" + conf.Host + "/" + conf.Database + "?charset=utf8&loc=Asia%2FShanghai"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(conf.MaxIdle)
	db.SetMaxOpenConns(conf.MaxOpen)
}

func Addr() string {
	return conf.PORT
}

func New(uuid uuid.UUID) (int64, error) {
	res, err := db.Exec("REPLACE INTO `seq_number` (uuid) VALUES (?)", uuid)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func NextId() int64 {
	s := atomic.LoadInt64(&startID)
	if s == 0 || curID >= maxID {
		mu.Lock()
		s = atomic.LoadInt64(&startID)
		if s == 0 || curID >= maxID {
			id, err := New(uid)
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
			return atomic.LoadInt64(&curID)
		}

		log.Printf("incr current id: %d failed", curID)
	}
}