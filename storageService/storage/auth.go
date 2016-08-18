package storage

import (
	"github.com/kataras/iris/utils"
	"gopkg.in/redis.v4"
)

func getSessionValue(redisCli *redis.ClusterClient, key string) (map[string]interface{}, error) {
	b, err := redisCli.Get(key).Bytes()
	if err != nil {
		return nil, err
	}
	r := make(map[string]interface{})
	err = utils.DeserializeBytes(b, &r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
