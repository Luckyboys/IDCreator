package Service

import (
	"bytes"
	"encoding/json"
	"github.com/Luckyboys/IDCreator/Common"
	"io"
)

type Message struct {
	Key            string
	IncrementValue uint64
}

func decode(content []byte) (Message, bool) {
	buf := bytes.NewReader(content)
	decoder := json.NewDecoder(buf)

	var m Message
	m.IncrementValue = 1
	m.Key = ""
	if err := decoder.Decode(&m); err == io.EOF {
		return m, false
	} else if err != nil {
		Common.CheckError(err, Common.ERROR)
		return m, false
	}

	return m, true
}
