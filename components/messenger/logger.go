package messenger

import (
	"fmt"
	"github.com/GabeCordo/toolchain/files"
	"log"
	"os"
	"time"
)

func GenerateFileName(endpoint string) (name string) {
	currTime := time.Now()
	currTimeStr := currTime.Format("2006:01:02_14:04:05")

	name = fmt.Sprintf("%s_%s.log", endpoint, currTimeStr)
	return name
}

func SaveToFile(dirPath, endpoint string, logs []string) bool {

	if _, err := os.Stat(dirPath); err != nil {
		log.Println("warning: cannot save logs to file, the save directory doesn't exist")
		return false
	}

	fileName := GenerateFileName(endpoint)
	path := files.EmptyPath().Dir(dirPath).File(fileName)

	file, err := path.Create()
	if err != nil {
		log.Println(err)
		return false
	}
	defer file.Close()

	for _, log := range logs {
		file.WriteString(log + "\n")
	}

	return true
}
