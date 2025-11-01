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
var NotFound = errors.New("a contact with the same identifier cannot be found")

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

func (contact *Contact) GetSubscriptions() []Subscription {

	contact.mutex.RLock()
	defer contact.mutex.RUnlock()

	copyOfSubscriptions := make([]Subscription, len(contact.Subscriptions))
	copy(contact.Subscriptions, copyOfSubscriptions)

	return copyOfSubscriptions
}

func (contact *Contact) AddSubscription(newSubscription Subscription) error {

	contact.mutex.Lock()
	defer contact.mutex.Unlock()

	for _, existingSubscription := range contact.Subscriptions {
		if existingSubscription == newSubscription {
			return DuplicateSubscription
		}
	}

	contact.Subscriptions = append(contact.Subscriptions, newSubscription)
	return nil
}

func (contact *Contact) RemoveSubscription(deleteSubscription Subscription) error {

	contact.mutex.Lock()
	defer contact.mutex.Unlock()

	contactSubscribed := false
	for idx, existingSubscription := range contact.Subscriptions {
		if existingSubscription == deleteSubscription {
			contact.Subscriptions = append(contact.Subscriptions[:idx], contact.Subscriptions[idx+1:]...)
			contactSubscribed = true
			break
		}
	}

	if !contactSubscribed {
		return UnknownSubscription
	} else {
		return nil
	}
}

func (contact *Contact) GetEmails() []Email {

	contact.mutex.RLock()
	defer contact.mutex.RUnlock()

	copyOfEmail := make([]Email, len(contact.Emails))
	copy(contact.Emails, copyOfEmail)

	return copyOfEmail
}

func (contact *Contact) AddEmail(newEmail Email) error {

	contact.mutex.Lock()
	defer contact.mutex.Unlock()

	for _, existingEmails := range contact.Emails {
		if existingEmails.value == newEmail.value {
			return DuplicateEmail
		}
	}

	contact.Emails = append(contact.Emails, newEmail)
	return nil
}

func (contact *Contact) RemoveEmail(deleteEmail Email) error {

	contact.mutex.Lock()
	defer contact.mutex.Unlock()

	contactHasThisEmail := false
	for idx, existingEmail := range contact.Emails {
		if existingEmail == deleteEmail {
			contact.Emails = append(contact.Emails[:idx], contact.Emails[idx+1:]...)
			contactHasThisEmail = true
			break
		}
	}

	if !contactHasThisEmail {
		return UnknownEmail
	} else {
		return nil
	}
}

type Database interface {
	GetContacts() []*Contact
	AddContact(newContact *Contact) error
	DeleteContact(identifier string) error
	Print()
}
