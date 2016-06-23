package generator

import (
	"sync/atomic"
	"time"
)

const UserTypeId = 1

//setp数等于机器数
const step = 1

//为每一种type分配一个计数器,初始值为机器Id
//type range : from 0 to 99
var counters [100]int64

//机器Id,从0开始递增分配,每1万一个轮回
var machineId int64 = 0

func init() {
	for i := range counters {
		counters[i] = machineId
	}

}

func Generate(typeid int64) int64 {
	//共19位,前13位是时间戳，中间4位是计数器，后2位是类型Id
	return (int64(time.Now().UnixNano())/1000000)*1000000 + (atomic.AddInt64(&counters[typeid], step)%10000)*100 + typeid
}
