package messenger

import (
	"fmt"
	"os"
	"testing"
)

func TestMessenger(t *testing.T) {

	_, isTestLocal := os.LookupEnv("MANGO_LOCAL_TEST")
	if !isTestLocal {
		t.Skip("this test can not be run on actions")
	}

	m := NewMessenger(true, true).SetupSMTP(DefaultEndpoint, Credentials{"~", "~"})

	m.Log("test", "this is a hello message")
	m.Log("test", "this is a hello message two")

	if success := m.Complete("gabecofficial@gmail.com"); !success {
		t.Error("email did not send")
	}
}

func TestGenerateFileName(t *testing.T) {
	name := GenerateFileName("vector")
	fmt.Println(name)
}

func TestSaveToFile(t *testing.T) {

	if success := SaveToFile("/Users/gabecordovado/Desktop/ETLSentiment/.logs", "test", []string{
		"log test 1",
		"log test 2",
	}); !success {
		t.Error("save to file failed")
	}

}
