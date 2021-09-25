package baijiayun

import (
	"context"
	"encoding/json"
	"task_client/utils/logger"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
)

const (
	IDKEY    = "task:id"
	IDVALUES = "task:class"
)

var rctx = context.Background()

type Bqueue struct {
}

func init() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       8,  // use default DB
	})

	_, err = Redis.Ping(rctx).Result()
	if err != nil {
		logger.Error("Redis Conn Error", err)
		panic(err)
	}
}

var Redis *redis.Client

func New() Queue {
	return &Bqueue{}
}

func (t *Bqueue) Serve() {
	t.Start()
}

func (t *Bqueue) Start() {

}

func (t *Bqueue) Push(v interface{}) int64 {
	s, _ := json.Marshal(v)
	id := t.getId()
	_, err := Redis.RPush(rctx, IDKEY, id).Result()
	logger.PanicError(err, "RPush", false)
	_, err = Redis.HMSet(rctx, IDVALUES, map[string]interface{}{
		cast.ToString(id): s,
	}).Result()

	logger.PanicError(err, "HMSet", false)

	return id
}

func (t *Bqueue) Remove(id int64) {
	Redis.HDel(rctx, IDVALUES, cast.ToString(id))
}

func (t *Bqueue) Pop() int64 {
	id, err := Redis.LPop(rctx, IDKEY).Int64()
	if err != nil {
		//logger.Error("Redis Prop", err.Error())
		return 0
	}
	return id
}

func (t *Bqueue) Get(id int64) string {
	r, err := Redis.HGet(rctx, IDVALUES, cast.ToString(id)).Result()
	if err != nil {
		logger.Error("Redis Prop", err.Error())
		return ""
	}
	return r
}

func (t *Bqueue) getId() int64 {
	return time.Now().Unix()
}
