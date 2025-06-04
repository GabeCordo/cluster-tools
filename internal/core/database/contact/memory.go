package contact

import "sync"

type Database struct {
	contacts map[string]*Contact
	mutex    sync.RWMutex
}

func New() *Database {
	directory := new(Database)
	directory.contacts = make(map[string]*Contact)
	return directory
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

func (db *Database) Save(path string) error {
	// TODO - not important feature currently
	return nil
}

func (db *Database) Load(path string) error {
	// TODO - not important feature currently
	return nil
}

func (db *Database) GetContacts() []*Contact {

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	listOfContact := make([]*Contact, len(db.contacts))

	idx := 0
	for _, contact := range db.contacts {
		listOfContact[idx] = contact
		idx++
	}

	return listOfContact
}

func (db *Database) AddContact(newContact *Contact) error {

	if newContact == nil {
		return NilContact
	}

	db.mutex.RLock()

	if _, found := db.contacts[newContact.Identifier]; found {
		db.mutex.RUnlock()
		return DuplicateContact
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
		return NotFound
	}

	db.mutex.RUnlock()

	db.mutex.Lock()
	defer db.mutex.Unlock()

	delete(db.contacts, identifier)
	return nil
}
