package storage

import (
	"github.com/kataras/iris/v12/sessions"
	"gopkg.in/redis.v4"
)

func getSessionValue(redisCli *redis.ClusterClient, key string) (map[string]interface{}, error) {
	b, err := redisCli.Get(key).Bytes()
	if err != nil {
		return nil, err
	}

	r := make(map[string]interface{})
	err = sessions.DefaultTranscoder.Unmarshal(b,&r)
	 
	 return r, err
}