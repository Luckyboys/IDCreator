package Service

import (
	"fmt"
	"github.com/Luckyboys/IDCreator/Common"
	"net"
)

func StartToListen() {
	const ip = "127.0.0.1"
	const port = 5354
	var address string = fmt.Sprintf("%s:%d", ip, port)
	Common.WriteLog(fmt.Sprintf("Starting to listen %s", address), Common.NOTICE)
	listener, err := net.Listen("tcp", "127.0.0.1:5354")
	if Common.CheckError(err, Common.ERROR) {
		return
	}

	defer listener.Close()

	initDB()

	for {
		conn, err := listener.Accept()
		if Common.CheckError(err, Common.WARNING) {
			continue
		}

		go handleConnection(conn)
	}
}
