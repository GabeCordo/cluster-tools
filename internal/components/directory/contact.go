package directory

import "errors"

var DuplicateSubscription = errors.New("subscription already exist")
var UnknownSubscription = errors.New("contact does not have this subscription")

var DuplicateEmail = errors.New("email already exist")
var UnknownEmail = errors.New("contact does not have this email")

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
