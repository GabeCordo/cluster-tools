package messenger

import (
	"fmt"
	"testing"
)

func TestMessenger(t *testing.T) {
	//ukfehuywjmkmnydc
	m := NewMessenger(DefaultEndpoint, Credentials{"~", "~"})

	m.Log("test", "this is a hello message")
	m.Log("test", "this is a hello message two")

	if success := m.Complete("test", []string{"gabecofficial@gmail.com"}); !success {
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
