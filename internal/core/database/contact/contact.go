package contact

import (
	"errors"
	"regexp"
	"sync"
)

var DuplicateSubscription = errors.New("subscription already exist")
var UnknownSubscription = errors.New("contact does not have this subscription")

var DuplicateEmail = errors.New("email already exist")
var UnknownEmail = errors.New("contact does not have this email")

var NilContact = errors.New("received contact was nil")
var DuplicateContact = errors.New("a contact with the same identifier already exists")
var ContactNotFound = errors.New("a contact with the same identifier cannot be found")

var InvalidEmailFormat = errors.New("the value held within the instance is not a valid email")

type Email struct {
	value string
}

var regexForEmailValidation, _ = regexp.Compile("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$")

func (email Email) Valid() bool {
	return regexForEmailValidation.MatchString(email.value)
}

func NewEmail(value string) (Email, error) {
	email := Email{value: value}

	if !email.Valid() {
		return Email{}, InvalidEmailFormat
	} else {
		return email, nil
	}
}

type Subscription uint16

const (
	Events Subscription = iota
	Logs
)

type Contact struct {
	Identifier    string         `json:"identifier"`
	Emails        []Email        `json:"receiver"`
	Subscriptions []Subscription `json:"subscriptions"`

	mutex sync.RWMutex
}

func NewContact(identifier string) *Contact {
	contact := new(Contact)

	contact.Identifier = identifier
	contact.Emails = make([]Email, 0)
	contact.Subscriptions = make([]Subscription, 0)

	return contact
}
