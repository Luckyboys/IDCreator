package Service

import (
	"fmt"
	"github.com/Luckyboys/IDCreator/Common"
	"github.com/bradfitz/gomemcache/memcache"
	"time"
)

type MemcacheClient struct {
	client *memcache.Client
	isInit bool
	lock   chan int
}

var instanceMemcacheClient = new(MemcacheClient)

func GetMemcacheClient() *MemcacheClient {
	if !instanceMemcacheClient.isInit {
		instanceMemcacheClient.client = memcache.New(Common.GetConfigInstance().Get("memcache", "127.0.0.1:11211"))
		instanceMemcacheClient.isInit = true
		instanceMemcacheClient.lock = make(chan int, 5)
		instanceMemcacheClient.lock <- 1
	}

	return instanceMemcacheClient
}

func (this *MemcacheClient) Incrment(key string, incrementValue uint64) uint64 {
	this.getLock()
	defer this.unlock()
	Common.GetLogger().WriteLog("Try to increment: "+key+" , "+fmt.Sprintf("%d", incrementValue), Common.NOTICE)
	newValue, err := instanceMemcacheClient.client.Increment(key, incrementValue)

	if Common.GetLogger().CheckError(err, Common.ERROR) {
		return 0
	}

	return newValue
}

func (this *MemcacheClient) Set(key string, value string) {
	this.getLock()
	defer this.unlock()
	Common.GetLogger().WriteLog("Try to set: "+key+" , "+value, Common.NOTICE)
	item := new(memcache.Item)
	item.Key = key
	item.Value = []byte(value)
	item.Expiration = int32(time.Now().AddDate(0, 0, 30).Unix())
	err := this.client.Set(item)

	if Common.GetLogger().CheckError(err, Common.ERROR) {
		return
	}
}

func (this *MemcacheClient) Get(key string) string {
	this.getLock()
	defer this.unlock()
	Common.GetLogger().WriteLog("Try to get: "+key, Common.NOTICE)
	item, err := this.client.Get(key)
	if Common.GetLogger().CheckError(err, Common.ERROR) {
		return ""
	}

	return string(item.Value)
}

func (this *MemcacheClient) unlock() {
	this.lock <- 1
}

func (this *MemcacheClient) getLock() {
	<-this.lock
}
