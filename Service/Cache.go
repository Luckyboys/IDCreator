package Service

import (
	"fmt"
	"github.com/Luckyboys/IDCreator/Common"
	"github.com/bradfitz/gomemcache/memcache"
	"strconv"
	"strings"
	"time"
)

var cachePostfix string = ""

type MemcacheClient struct {
	client  *memcache.Client
	isUsing bool
	lock    chan int
}

type MemcachePool struct {
	clients       []*MemcacheClient
	locker        chan int
	isInit        bool
	syncSearching chan int
}

var instanceMemcachePool = new(MemcachePool)

func GetMemcacheClient() *MemcacheClient {
	return instanceMemcachePool.getClient()
}

func (this *MemcachePool) init() {
	this.isInit = true
	threadCount := Common.GetConfigInstance().GetUint("memcachethreadcount", 1)
	Common.GetLogger().WriteLog(fmt.Sprintf("Start to connect memcache , %d", threadCount), Common.NOTICE)
	cachePostfix = Common.GetConfigInstance().Get("cachepostfix", "")
	this.locker = make(chan int, threadCount)

	this.clients = make([]*MemcacheClient, threadCount)

	for i := 0; uint64(i) < threadCount; i++ {
		this.clients[i] = new(MemcacheClient)
		this.clients[i].client = memcache.New(Common.GetConfigInstance().Get("memcache", "127.0.0.1:11211"))
		this.clients[i].isUsing = false
		this.clients[i].lock = make(chan int, 1)
		this.clients[i].lock <- 1
		this.locker <- 1
	}
	this.syncSearching = make(chan int, 1)
	this.syncSearching <- 1
}

func (this *MemcachePool) getClient() *MemcacheClient {

	if !this.isInit {
		this.init()
	}

	this.getLock()

	<-this.syncSearching

	for {
		for _, client := range this.clients {
			if client.isUsing {
				continue
			}

			client.isUsing = true
			this.syncSearching <- 1
			return client
		}

		time.Sleep(10 * time.Microsecond)
	}

}

func (this *MemcacheClient) Free() {
	this.isUsing = false
	instanceMemcachePool.unlock()
}

func (this *MemcacheClient) Incrment(key string, incrementValue uint64) uint64 {
	this.getLock()
	defer this.unlock()

	Common.GetLogger().WriteLog("Try to increment: "+this.getRealKey(key)+" , "+fmt.Sprintf("%d", incrementValue), Common.NOTICE)
	newValue, err := this.client.Increment(this.getRealKey(key), incrementValue)

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

func (this *MemcachePool) unlock() {
	this.locker <- 1
}

func (this *MemcachePool) getLock() {
	<-this.locker
}

func (this *MemcacheClient) getRealKey(key string) string {
	return strings.Join([]string{key, cachePostfix}, "")
}
