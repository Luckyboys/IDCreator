package Service

import (
	"fmt"
	"github.com/Luckyboys/IDCreator/Common"
	"github.com/bradfitz/gomemcache/memcache"
	"strconv"
	"strings"
)

var cachePostfix string = ""

type MemcacheClient struct {
	client *memcache.Client
	isInit bool
	lock   chan int
}

type MemcachePool struct {
	clients []MemcacheClient
	locker  chan int
	isInit  bool
}

var instanceMemcacheClient = new(MemcacheClient)

func GetMemcacheClient() *MemcacheClient {
	if !instanceMemcacheClient.isInit {
		instanceMemcacheClient.client = memcache.New(Common.GetConfigInstance().Get("memcache", "127.0.0.1:11211"))
		instanceMemcacheClient.isInit = true
		instanceMemcacheClient.lock = make(chan int, 1)
		instanceMemcacheClient.lock <- 1
		cachePostfix = Common.GetConfigInstance().Get("cachepostfix", "")
	}

	return instanceMemcacheClient
}

func (this *MemcachePool) init() {
	var threadCount uint64 = strconv.ParseUint(Common.GetConfigInstance().Get("memcachethreadcount", "10"), 10, 64)

	this.locker = make(chan int, threadCount)

	for i := 0; i < threadCount; i++ {
		this.clients[i] = new(MemcacheClient)

	}
	this.isInit = true
}

func (this *MemcachePool) getClient() MemcacheClient {

	if !this.isInit {
		this.init()
	}

	for _, client := range this.clients {
		client.isInit
	}
}

func (this *MemcacheClient) Incrment(key string, incrementValue uint64) uint64 {
	this.getLock()
	defer this.unlock()

	Common.GetLogger().WriteLog("Try to increment: "+this.getRealKey(key)+" , "+fmt.Sprintf("%d", incrementValue), Common.NOTICE)
	newValue, err := instanceMemcacheClient.client.Increment(this.getRealKey(key), incrementValue)

	if Common.GetLogger().CheckError(err, Common.ERROR) {
		return 0
	}

	return newValue
}

func (this *MemcacheClient) Set(key string, value string) {
	this.getLock()
	defer this.unlock()
	Common.GetLogger().WriteLog(fmt.Sprintf("Try to set: %s , %s , %b", this.getRealKey(key), value, []byte(value)), Common.NOTICE)
	item := new(memcache.Item)
	item.Key = this.getRealKey(key)
	item.Value = []byte(value)
	expire, _ := strconv.Atoi(Common.GetConfigInstance().Get("memcacheexpire", strconv.FormatInt(15*86400, 32)))
	item.Expiration = int32(expire)
	Common.GetLogger().WriteLog(fmt.Sprintf("Item: %s", item), Common.NOTICE)

	err := this.client.Set(item)

	if Common.GetLogger().CheckError(err, Common.ERROR) {
		return
	}
}

func (this *MemcacheClient) Get(key string) string {
	this.getLock()
	defer this.unlock()
	Common.GetLogger().WriteLog(fmt.Sprintf("Try to get: %s , %b ", this.getRealKey(key), []byte(this.getRealKey(key))), Common.NOTICE)
	item, err := this.client.Get(this.getRealKey(key))
	if Common.GetLogger().CheckError(err, Common.ERROR) {
		return ""
	}

	Common.GetLogger().WriteLog(fmt.Sprintf("Get result: %s", item.Value), Common.NOTICE)

	return string(item.Value)
}

func (this *MemcacheClient) unlock() {
	this.lock <- 1
}

func (this *MemcacheClient) getLock() {
	<-this.lock
}

func (this *MemcacheClient) getRealKey(key string) string {
	return strings.Join([]string{key, cachePostfix}, "")
}
