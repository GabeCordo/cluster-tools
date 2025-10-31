package in_memory

import (
	"github.com/FortifiedCode/flock/internal/core/database/contact"
	"sync"
)

type Database struct {
	contacts map[string]*contact.Contact
	mutex    sync.RWMutex
}

func New() *Database {
	directory := new(Database)
	directory.contacts = make(map[string]*contact.Contact)
	return directory
}

func (db *Database) GetContacts() []*contact.Contact {

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	listOfContact := make([]*contact.Contact, len(db.contacts))

	idx := 0
	for _, c := range db.contacts {
		listOfContact[idx] = c
		idx++
	}

	return listOfContact
}

func (db *Database) AddContact(newContact *contact.Contact) error {

	if newContact == nil {
		return contact.NilContact
	}

	db.mutex.RLock()

	if _, found := db.contacts[newContact.Identifier]; found {
		db.mutex.RUnlock()
		return contact.DuplicateContact
	}

	db.mutex.RUnlock()
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.contacts[newContact.Identifier] = newContact // transfer ownership

	return nil
}

func (db *Database) DeleteContact(identifier string) error {

	db.mutex.RLock()

	if _, found := db.contacts[identifier]; !found {
		db.mutex.RUnlock()
		return contact.NotFound
	}

	db.mutex.RUnlock()

	db.mutex.Lock()
	defer db.mutex.Unlock()

	delete(db.contacts, identifier)
	return nil
}

func (db *Database) Print() {

}
