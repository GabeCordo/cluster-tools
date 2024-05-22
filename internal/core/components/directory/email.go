package directory

import (
	"errors"
	"regexp"
)

var InvalidEmailFormat = errors.New("the value held within the instance is not a valid email")

var regexForEmailValidation, _ = regexp.Compile("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")

func (email Email) Valid() bool {
	return regexForEmailValidation.MatchString(email.value)
}
