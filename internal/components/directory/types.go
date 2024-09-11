package directory

import "sync"

type Email struct {
	value string
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

type Directory struct {
	contacts map[string]*Contact

	mutex sync.RWMutex
}

func NewDirectory() *Directory {
	directory := new(Directory)
	directory.contacts = make(map[string]*Contact)
	return directory
}
