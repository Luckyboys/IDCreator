package Common

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

type Config struct {
	data map[string]string
}

var instanceConfig *Config = new(Config)

func (this *Config) Init(configFilePath string) {
	this.data = make(map[string]string)

	file, err := os.Open(configFilePath)
	defer file.Close()

	if err != nil {
		fmt.Println(fmt.Sprintf("Load ConfigError => %s", err))
	}

	data := make([]byte, 1024)
	var configBuffer bytes.Buffer

	for {
		count, err := file.Read(data)
		if err != io.EOF && err != nil {

			fmt.Println(fmt.Sprintf("Load ConfigError => %s", err))
			return
		}

		if count == 0 {
			break
		}

		configBuffer.Write(data[:count])
	}

	var configString string = strings.Replace(string(configBuffer.Bytes()), "\r", "\n", -1)

	var lines []string = strings.Split(configString, "\n")
	for lineN := 0; lineN < len(lines); lineN++ {
		lines[lineN] = strings.Trim(lines[lineN], " ")
		if strings.Index(lines[lineN], ";") == 0 || strings.Index(lines[lineN], "#") == 0 {
			continue
		}

		var line []string = strings.Split(lines[lineN], "=")
		if len(line) != 2 {
			continue
		}
		line[0] = strings.Trim(line[0], " ")
		line[1] = strings.Trim(line[1], " ")

		if len(line[0]) == 0 {
			continue
		}

		this.data[line[0]] = line[1]
	}
}

func (this *Config) PrintData() {

	for key, value := range this.data {
		GetLogger().WriteLog(fmt.Sprintf("%s = %s\n", key, value), NOTICE)
	}

}

func (this *Config) Get(key string, defaultValue string) string {
	if _, exist := this.data[key]; !exist {
		return defaultValue
	}
	return this.data[key]
}

func GetConfigInstance() *Config {
	return instanceConfig
}
