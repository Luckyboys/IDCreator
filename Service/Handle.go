package Service

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/Luckyboys/IDCreator/Common"
	"io"
	"net"
	"strconv"
	"time"
)

const CONTENT_LENGTH_SIZE = 4

func handleConnection(connection net.Conn) {
	for {
		contentLength := _getMessageLength(connection)
		defer connection.Close()
		Common.GetLogger().WriteLog(fmt.Sprintf("Known Content Length: %d", contentLength), Common.NOTICE)
		if contentLength <= 0 {
			return
		}
		contentBuffer, readSuccess := _read(connection, contentLength, 500*time.Millisecond)
		Common.GetLogger().WriteLog(fmt.Sprintf("Read Content: %s", bytes.NewBuffer(contentBuffer).String()), Common.NOTICE)
		if !readSuccess {
			return
		}

		message, decodeSuccess := decode(contentBuffer)
		if decodeSuccess == false {
			return
		}

		Common.GetLogger().WriteLog("Decode Successed", Common.NOTICE)
		var value uint64 = 0
		switch action := message.Action; action {
		case ACTION_GET:
			value = GetKeyBoxInstance().get(message.Key)
			break
		case ACTION_INCR:
			value = GetKeyBoxInstance().incr(message.Key, message.IncrementValue)
			break
		}

		Common.GetLogger().WriteLog(fmt.Sprintf("Increment OK: key => %s , value => %d", message.Key, value), Common.NOTICE)
		_write(connection, "{\"result\":\""+strconv.FormatUint(value, 10)+"\"}")
		Common.GetLogger().WriteLog("Content Sent , Time to close", Common.NOTICE)
	}
}

func _getMessageLength(connection net.Conn) uint32 {
	var contentLength uint32
	contentLengthBuffer, readSuccess := _read(connection, CONTENT_LENGTH_SIZE, 500*time.Millisecond)

	if !readSuccess {
		return 0
	}
	buf := bytes.NewReader(contentLengthBuffer)

	binary.Read(buf, binary.BigEndian, &contentLength)
	return contentLength
}

func _read(connection net.Conn, length uint32, timeoutNano time.Duration) ([]byte, bool) {

	buf := bytes.NewBuffer(make([]byte, 0))

	var needGetLength uint32 = length
	_markStartTime()

	for uint32(buf.Len()) <= length && _reachTimeoutLimit(timeoutNano) {
		var contentLengthBuffer []byte = make([]byte, needGetLength)
		iLen, err := connection.Read(contentLengthBuffer)

		if err == io.EOF {
			continue
		}

		if Common.GetLogger().CheckError(err, Common.WARNING) {
			return nil, false
		}

		buf.Write(contentLengthBuffer)
		needGetLength -= uint32(iLen)
	}

	if uint32(buf.Len()) != length {
		Common.GetLogger().WriteLog("ContentLength Error", Common.ERROR)
		return nil, false
	}

	return buf.Bytes(), true
}

func _write(connection net.Conn, message string) {

	var content = bytes.NewBufferString(message)
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint32(content.Len()))
	if Common.GetLogger().CheckError(err, Common.ERROR) {
		return
	}

	connection.Write(buf.Bytes())
	connection.Write(content.Bytes())

	Common.GetLogger().WriteLog("Writed", Common.NOTICE)
}

var startReadTime int64

func _markStartTime() {
	startReadTime = time.Now().UnixNano()
}

func _reachTimeoutLimit(timeoutNano time.Duration) bool {
	return !(time.Now().UnixNano()-startReadTime > int64(timeoutNano))
}
