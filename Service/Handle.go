package Service

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/Luckyboys/IDCreator/Common"
	"io"
	"net"
	"strconv"
)

const CONTENT_LENGTH_SIZE = 4

func handleConnection(connection net.Conn) {
	contentLength := _getMessageLength(connection)
	defer connection.Close()
	Common.WriteLog(fmt.Sprintf("Known Content Length: %d", contentLength), Common.NOTICE)
	if contentLength <= 0 {
		return
	}
	contentBuffer, readSuccess := _read(connection, contentLength)
	Common.WriteLog(fmt.Sprintf("Read Content: %s", bytes.NewBuffer(contentBuffer).String()), Common.NOTICE)
	if !readSuccess {
		return
	}

	message, decodeSuccess := decode(contentBuffer)
	if decodeSuccess == false {
		return
	}

	Common.WriteLog("Decode Successed", Common.NOTICE)
	value := incr(message.Key, message.IncrementValue)
	Common.WriteLog(fmt.Sprintf("Increment OK: key => %s , value => %d", message.Key, value), Common.NOTICE)
	_write(connection, "{\"result\":\""+strconv.FormatUint(value, 10)+"\"}")
	Common.WriteLog("Content Sent , Time to close", Common.NOTICE)
}

func _getMessageLength(connection net.Conn) uint32 {
	var contentLength uint32
	contentLengthBuffer, readSuccess := _read(connection, CONTENT_LENGTH_SIZE)

	if !readSuccess {
		return 0
	}
	buf := bytes.NewReader(contentLengthBuffer)

	binary.Read(buf, binary.BigEndian, &contentLength)
	return contentLength
}

func _read(connection net.Conn, length uint32) ([]byte, bool) {
	var contentLengthBuffer []byte = make([]byte, length)
	iLen, err := connection.Read(contentLengthBuffer)

	if err == io.EOF {
		return nil, false
	}

	if Common.CheckError(err, Common.WARNING) {
		return nil, false
	}

	if uint32(iLen) != length {
		Common.WriteLog("ContentLength Error", Common.ERROR)
		return nil, false
	}

	return contentLengthBuffer, true
}

func _write(connection net.Conn, message string) {

	var content = bytes.NewBufferString(message)
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint32(content.Len()))
	if Common.CheckError(err, Common.ERROR) {
		return
	}

	connection.Write(buf.Bytes())
	connection.Write(content.Bytes())

	Common.WriteLog("Writed", Common.NOTICE)
}
