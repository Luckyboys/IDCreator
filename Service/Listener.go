package Service

import (
	"fmt"
	"github.com/Luckyboys/IDCreator/Common"
	"net"
)

func StartToListen() {
	ip := Common.GetConfigInstance().Get("listenhost", "127.0.0.1")
	port := Common.GetConfigInstance().Get("listenport", "5354")
	var address string = fmt.Sprintf("%s:%s", ip, port)
	Common.GetLogger().WriteLog(fmt.Sprintf("Starting to listen %s", address), Common.NOTICE)
	listener, err := net.Listen("tcp", address)
	if Common.GetLogger().CheckError(err, Common.ERROR) {
		return
	}

	defer listener.Close()

	initDB()

	for {
		Common.Try(
			func() {
				conn, err := listener.Accept()
				if Common.GetLogger().CheckError(err, Common.WARNING) {
					return
				}

				go handleConnection(conn)
			},
			func(e interface{}) {
				Common.GetLogger().WriteLog(fmt.Sprintf("Listener ErrorMessage: %s", e), Common.ERROR)
			})

	}
}
