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
	mu      sync.Mutex
)

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

func New(uuid uuid.UUID) (int64, error) {
	res, err := db.Exec("REPLACE INTO `seq_number` (uuid) VALUES (?)", uuid)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

// 生成新的id
func genId() {
	mu.Lock()
	s := atomic.LoadInt64(&startID)
	if s == 0 || curID >= maxID {
		id, err := New(uid)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("get new id : %d for ip : %s，uuid : %s\n", id, host, uid)

		atomic.StoreInt64(&startID, id*conf.STEP)

		// curID如果大于等于maxID则不考虑更新curID
		if s == 0 {
			atomic.StoreInt64(&curID, id*conf.STEP)
		}

		atomic.StoreInt64(&maxID, (id+1)*conf.STEP)
	}

	mu.Unlock()
}

func NextId() int64 {
	// 获取起点值
	s := atomic.LoadInt64(&startID)

	// 这里缓存curID
	nextId := atomic.LoadInt64(&curID)

	// 这里进行初步判断
	if s == 0 || nextId >= maxID {
		genId()
	}

	// 有点暴力，需要压力测试看看效果
	for {
		// 需要保证nextId不能大于maxID，防止出现数据重复
		nextId = atomic.LoadInt64(&curID)

		// 这里的curID在高并发的情况下，可能大于maxID
		if nextId >= maxID {
			genId()
		}

		if atomic.CompareAndSwapInt64(&curID, nextId, nextId+1) {
			// quit only when CompareAndSwap success, otherwise retry
			return nextId
		}

		log.Printf("incr current id: %d failed", nextId)
	}
}