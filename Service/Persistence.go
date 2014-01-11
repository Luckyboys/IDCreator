package Service

import (
	"database/sql"
	// "fmt"
	"github.com/Luckyboys/IDCreator/Common"
	"github.com/Luckyboys/StringBuilder"
	_ "github.com/go-sql-driver/mysql"
)

var mx = make(chan int, 5)

//TODO 连接别连来连去，用完就Hold住。连接池维护
func initDB() {
	mx <- 1
	mx <- 1
	mx <- 1
	mx <- 1
	mx <- 1
}

func getDBValue(key string) uint64 {
	<-mx
	var value uint64 = 0

	db, err := sql.Open("mysql", getConnectMySQLString())

	if Common.GetLogger().CheckError(err, Common.ERROR) {
		mx <- 1
		return 0
	}

	defer db.Close()

	statmentSelect, err := db.Prepare("SELECT `value` FROM `counter` WHERE `key` = ?")
	if Common.GetLogger().CheckError(err, Common.ERROR) {
		mx <- 1
		return 0
	}

	result, err := statmentSelect.Query(key)

	if Common.GetLogger().CheckError(err, Common.ERROR) {
		mx <- 1
		return 0
	}

	for result.Next() {
		result.Scan(&value)
		break
	}

	mx <- 1
	return value
}

func setDBValue(key string, value uint64) {
	<-mx
	db, err := sql.Open("mysql", getConnectMySQLString())

	if Common.GetLogger().CheckError(err, Common.ERROR) {
		mx <- 1
		return
	}

	defer db.Close()

	statmentInsert, err := db.Prepare("INSERT INTO `counter` ( `key` , `value` ) VALUES ( ? , ? ) ON DUPLICATE KEY UPDATE `value` = ?")
	if Common.GetLogger().CheckError(err, Common.ERROR) {
		mx <- 1
		return
	}

	result, err := statmentInsert.Exec(key, value, value)

	if Common.GetLogger().CheckError(err, Common.ERROR) {
		mx <- 1
		return
	}

	affectedRowCount, err := result.RowsAffected()

	if affectedRowCount <= 0 {
		Common.GetLogger().WriteLog("Can't Save key: "+key+" , value at: "+string(value), Common.ERROR)
	}
	mx <- 1
}

func getConnectMySQLString() string {

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
