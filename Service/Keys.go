package Service

import (
	"fmt"
	"github.com/Luckyboys/IDCreator/Common"
	"strconv"
	"sync/atomic"
	"time"
)

type KeyBox struct {
	data          map[string]*uint64
	isUseMemcache bool
	isInit        bool
}

var instanceKeyBox *KeyBox = new(KeyBox)

func GetKeyBoxInstance() *KeyBox {
	if !instanceKeyBox.isInit {
		instanceKeyBox.init()
		instanceKeyBox.isInit = true
	}

	return instanceKeyBox
}

func (this *KeyBox) init() {
	this.data = make(map[string]*uint64)
	this.isUseMemcache = Common.GetConfigInstance().Get("isusememcache", "0") == "1"
}

func (this *KeyBox) incr(key string, incrementValue uint64) (value uint64) {

	if this.isUseMemcache {
		memcacheClient := GetMemcacheClient()
		defer memcacheClient.Free()
		value = memcacheClient.Incrment(key, incrementValue)
		Common.GetLogger().WriteLog("increment", Common.NOTICE)
		if value <= uint64(10000) {
			memcacheClient.Set(key, fmt.Sprintf("%d", this.getNewKeyValue(key)))
			value = memcacheClient.Incrment(key, incrementValue)
		}

	} else {
		if _, exist := this.data[key]; !exist {

			this.data[key] = new(uint64)
			atomic.StoreUint64(this.data[key], this.getNewKeyValue(key))
		}
		value = atomic.AddUint64(this.data[key], incrementValue)
	}

	if value%10 == 0 {
		go setDBValue(key, value)
	}

	if value%100 == 0 {
		Common.GetLogger().WriteLog(fmt.Sprintf("key: %s , value: %d", key, value), Common.INFO)
	}
	return
}

func (this *KeyBox) get(key string) (value uint64) {

	if this.isUseMemcache {
		memcacheClient := GetMemcacheClient()
		defer memcacheClient.Free()
		value, _ = strconv.ParseUint(memcacheClient.Get(key), 10, 64)
		if value <= 10000 {
			value = this.getNewKeyValue(key)
			memcacheClient.Set(key, fmt.Sprintf("%d", value))
		}
	} else {
		if _, exist := this.data[key]; !exist {

			this.data[key] = new(uint64)
			atomic.StoreUint64(this.data[key], this.getNewKeyValue(key))
		}
		value = *this.data[key]
	}
	return
}

func (this *KeyBox) getNewKeyValue(key string) uint64 {

	var returnValue uint64 = 0
	returnValue = getDBValue(key)
	Common.GetLogger().WriteLog(fmt.Sprintf("Get From DB key: %s , value: %d", key, returnValue), Common.NOTICE)
	if returnValue == 0 {
		returnValue = uint64(time.Now().Unix())
		go setDBValue(key, returnValue)
	} else {
		returnValue += 100
		go setDBValue(key, returnValue)
	}
	Common.GetLogger().WriteLog(fmt.Sprintf("Return Get From DB key: %s , value: %d", key, returnValue), Common.NOTICE)
	return returnValue
}
