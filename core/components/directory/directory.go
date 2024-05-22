package directory

import "errors"

var NilContact = errors.New("received contact was nil")
var DuplicateContact = errors.New("a contact with the same identifier already exists")
var ContactNotFound = errors.New("a contact with the same identifier cannot be found")

func (directory *Directory) Save(path string) error {
	// TODO - not important feature currently
	return nil
}

func (directory *Directory) Load(path string) error {
	// TODO - not important feature currently
	return nil
}

func (directory *Directory) GetContacts() []*Contact {

	directory.mutex.RLock()
	defer directory.mutex.RUnlock()

	listOfContact := make([]*Contact, len(directory.contacts))

	idx := 0
	for _, contact := range directory.contacts {
		listOfContact[idx] = contact
		idx++
	}

	return listOfContact
}

func (directory *Directory) AddContact(newContact *Contact) error {

	if newContact == nil {
		return NilContact
	}

	directory.mutex.RLock()

	if _, found := directory.contacts[newContact.Identifier]; found {
		directory.mutex.RUnlock()
		return DuplicateContact
	}

	directory.mutex.RUnlock()
	directory.mutex.Lock()
	defer directory.mutex.Unlock()

	directory.contacts[newContact.Identifier] = newContact // transfer ownership

	return nil
}

func (directory *Directory) DeleteContact(identifier string) error {

	directory.mutex.RLock()

	if _, found := directory.contacts[identifier]; !found {
		directory.mutex.RUnlock()
		return ContactNotFound
	}

	directory.mutex.RUnlock()

	directory.mutex.Lock()
	defer directory.mutex.Unlock()

	delete(directory.contacts, identifier)
	return nil
}
