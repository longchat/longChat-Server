package util

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func RandomString(num int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	var temp []byte = make([]byte, num)
	for i := 0; i < num; i++ {
		temp[i] = byte(33 + rand.Intn(93))
	}
	return string(temp[:])
}

func RandomInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func NewToken(id int64, privateToken string, expireIn time.Duration) string {
	idBytes := []byte(fmt.Sprintf("%d", id))
	idEn := base64.URLEncoding.EncodeToString(idBytes)
	expireBytes := []byte(fmt.Sprintf("%d", time.Now().UnixNano()+int64(expireIn)))
	expireEn := base64.URLEncoding.EncodeToString(expireBytes)
	sha := sha256.Sum256([]byte(idEn + expireEn + privateToken))
	return idEn + ":" + expireEn + ":" + base64.URLEncoding.EncodeToString(sha[:])
}

func DecodeToken(accessToken string, privateToken string) (id int64, expireAt int64, isValid bool) {
	tokens := strings.Split(accessToken, ":")
	if len(tokens) < 3 {
		return
	}
	cksum := sha256.Sum256([]byte(tokens[0] + tokens[1] + privateToken))
	if base64.URLEncoding.EncodeToString(cksum[:]) != tokens[2] {
		return
	}
	ids, err := base64.URLEncoding.DecodeString(tokens[0])
	if err != nil {
		return
	}
	expireAts, err := base64.URLEncoding.DecodeString(tokens[1])
	if err != nil {
		return
	}
	id, err = strconv.ParseInt(string(ids), 10, 64)
	if err != nil {
		return
	}
	expireAt, err = strconv.ParseInt(string(expireAts), 10, 64)
	if err != nil {
		return
	}
	isValid = true
	return
}
