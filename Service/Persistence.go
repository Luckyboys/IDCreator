package Service

import (
	"database/sql"
	"fmt"
	"github.com/Luckyboys/IDCreator/Common"
	"github.com/Luckyboys/StringBuilder"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type sqlClient struct {
	client    *sql.DB
	isUsing   bool
	lock      chan int
	tableName string
}

type DBPool struct {
	clients       []*sqlClient
	locker        chan int
	syncSearching chan int
	isInit        bool
}

var instanceDBPool = new(DBPool)

func initDB() {
	if !instanceDBPool.isInit {
		instanceDBPool.init()
	}
}

func (this *DBPool) init() {
	this.isInit = true
	threadCount := Common.GetConfigInstance().GetUint("dbthreadcount", 1)
	Common.GetLogger().WriteLog(fmt.Sprintf("Start to connect db , %d", threadCount), Common.NOTICE)

	this.locker = make(chan int, threadCount)

	this.clients = make([]*sqlClient, threadCount)

	for i := 0; uint64(i) < threadCount; i++ {
		this.clients[i] = new(sqlClient)
		db, err := sql.Open("mysql", this.getConnectMySQLString())
		this.clients[i].client = db

		if Common.GetLogger().CheckError(err, Common.ERROR) {
			panic("Error MySQL")
		}

		this.clients[i].isUsing = false
		this.clients[i].lock = make(chan int, 1)
		this.clients[i].lock <- 1
		this.clients[i].tableName = Common.GetConfigInstance().Get("tablename", "counter")
		this.locker <- 1
	}
	this.syncSearching = make(chan int, 1)
	this.syncSearching <- 1
}

func (this *DBPool) getClient() *sqlClient {

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

func (this *sqlClient) Free() {
	this.isUsing = false
	instanceDBPool.unlock()
}

func (this *DBPool) unlock() {
	this.locker <- 1
}

func (this *DBPool) getLock() {
	<-this.locker
}

func (this *DBPool) getConnectMySQLString() string {

	user := Common.GetConfigInstance().Get("user", "root")

	password := Common.GetConfigInstance().Get("password", "")

	dbname := Common.GetConfigInstance().Get("dbname", "test")

	host := Common.GetConfigInstance().Get("host", "127.0.0.1")

	port := Common.GetConfigInstance().Get("port", "3306")

	//"root@tcp(127.0.0.1:3306)/test"
	var connectString *StringBuilder.StringBuilder = StringBuilder.GetStringBuilder()

	connectString.Append(user)

	if password != "" {
		connectString.Append(":" + password)
	}
	connectString.Append("@tcp(" + host + ":" + port + ")/" + dbname)

	return connectString.String()
}

func getDBValue(key string) uint64 {

	var value uint64 = 0

	db := instanceDBPool.getClient()

	defer db.Free()

	statmentSelect, err := db.client.Prepare(fmt.Sprintf("SELECT `value` FROM `%s` WHERE `key` = ?", db.tableName))
	if Common.GetLogger().CheckError(err, Common.ERROR) {
		return 0
	}

	result, err := statmentSelect.Query(key)

	if Common.GetLogger().CheckError(err, Common.ERROR) {
		return 0
	}

	for result.Next() {
		result.Scan(&value)
		break
	}

	return value
}

func setDBValue(key string, value uint64) {

	db := instanceDBPool.getClient()

	defer db.Free()

	statmentInsert, err := db.client.Prepare(fmt.Sprintf("INSERT INTO `%s` ( `key` , `value` ) VALUES ( ? , ? ) ON DUPLICATE KEY UPDATE `value` = ?", db.tableName))
	if Common.GetLogger().CheckError(err, Common.ERROR) {
		return
	}

	result, err := statmentInsert.Exec(key, value, value)

	if Common.GetLogger().CheckError(err, Common.ERROR) {
		return
	}

	affectedRowCount, err := result.RowsAffected()

	if affectedRowCount <= 0 {
		Common.GetLogger().WriteLog("Can't Save key: "+key+" , value at: "+string(value), Common.ERROR)
	}
}
