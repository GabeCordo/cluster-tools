package template

import (
	"log"
	"time"
)

const (
	StartIdentifier  uint8  = '<'
	EndIdentifier    uint8  = '>'
	EmptyIdentifier  string = ""
	InPlaceholder    bool   = true
	NotInPlaceholder bool   = false
	TimeFormat       string = "2006-01-02 15:04:05"
)

func Process(raw []byte, match map[string]string) []byte {
	var processedString string

	stringRepOfBytes := string(raw)

	placeholderFlag := NotInPlaceholder

	var identifier string
	for c := range stringRepOfBytes {
		char := stringRepOfBytes[c]

		if placeholderFlag {
			if char == EndIdentifier {
				if identifier == "date" {
					processedString += time.Now().Format(TimeFormat)
				} else if replacement, found := match[identifier]; found {
					processedString += replacement
				} else {
					log.Println("warning: unknown identifier (" + identifier + ")")
				}

				identifier = EmptyIdentifier
				placeholderFlag = NotInPlaceholder
			} else {
				identifier += string(char)
			}
		} else {
			if char == StartIdentifier {
				placeholderFlag = InPlaceholder
			} else {
				processedString += string(char)
			}
		}
	}

	return []byte(processedString)
}
