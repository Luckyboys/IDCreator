package Service

import (
	"fmt"
	"github.com/Luckyboys/IDCreator/Common"
	"sync/atomic"
	"time"
)

var keys map[string]*uint64 = make(map[string]*uint64)

func incr(key string, incrementValue uint64) uint64 {

	var value uint64
	if _, exist := keys[key]; !exist {
		keys[key] = new(uint64)
		atomic.StoreUint64(keys[key], _getNewKeyValue(key))
	}

	value = atomic.AddUint64(keys[key], incrementValue)

	if value%10 == 0 {
		go setDBValue(key, value)
	}

	if value%100 == 0 {
		Common.WriteLog(fmt.Sprintf("key: %s , value: %d", key, value), Common.INFO)
	}

	return value
}

func _getNewKeyValue(key string) uint64 {

	var returnValue uint64 = 0
	returnValue = getDBValue(key)
	Common.WriteLog(fmt.Sprintf("Get From DB key: %s , value: %d", key, returnValue), Common.NOTICE)
	if returnValue == 0 {
		returnValue = uint64(time.Now().Unix())
		go setDBValue(key, returnValue)
	} else {
		returnValue += 100
		go setDBValue(key, returnValue)
	}
	Common.WriteLog(fmt.Sprintf("Return Get From DB key: %s , value: %d", key, returnValue), Common.NOTICE)
	return returnValue
}
