package storage

import (
	"github.com/kataras/iris/core/memstore"
	"github.com/kataras/iris/sessions"

	"gopkg.in/redis.v4"
)

func getSessionValue(redisCli *redis.ClusterClient, key string) (memstore.Store, error) {
	var emptyValues memstore.Store

	b, err := redisCli.Get(key).Bytes()
	if err != nil {
		return emptyValues, err
	}

	r, err := sessions.DecodeRemoteStore(b) // decode the whole value, as a remote store
	if err != nil {
		return emptyValues, err
	}

	return r.Values, nil
}
