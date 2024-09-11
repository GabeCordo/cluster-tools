package components

type Subscription uint16

type Email interface {
	Value() string
	Valid() bool
}

type Contact interface {
	GetSubscriptions() []Subscription
	AddSubscription(newSubscription Subscription) error
	RemoveSubscription(deleteSubscription Subscription) error
	GetEmails() []Email
	AddEmail(newEmail Email) error
	RemoveEmail(deleteEmail Email) error
}

type Directory interface {
	GetContacts() []*Contact
	AddContact(newContact *Contact) error
	DeleteContact(identifier string) error
}
