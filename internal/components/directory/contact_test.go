package directory

import "testing"

func TestContact_AddSubscription(t *testing.T) {

	contact := NewContact("test")

	if err := contact.AddSubscription(Events); err != nil {
		t.Error(err)
	}

	if err := contact.AddSubscription(Events); err == nil {
		t.Error("expected contact to reject duplicate subscription")
	}
}

func TestContact_RemoveSubscription(t *testing.T) {

	contact := NewContact("test")

	if err := contact.AddSubscription(Events); err != nil {
		t.Error(err)
	}

	if err := contact.AddSubscription(Logs); err != nil {
		t.Error(err)
	}

	if err := contact.RemoveSubscription(Events); err != nil {
		t.Error(err)
	}

	if subscriptions := contact.GetSubscriptions(); len(subscriptions) != 1 {
		t.Error("expected 1 subscription to remain")
	}
}

func TestContact_AddEmail(t *testing.T) {

	contact := NewContact("test")

	email, err := NewEmail("john.doe@gmail.com")
	if err != nil {
		t.Error(err)
		return
	}

	if err := contact.AddEmail(email); err != nil {
		t.Error(err)
	}

	if err := contact.AddEmail(email); err == nil {
		t.Error("expected contact to reject duplicate subscription")
	}
}

func TestContact_RemoveEmail(t *testing.T) {

	contact := NewContact("test")

	email1, _ := NewEmail("john.doe@gmail.com")
	email2, _ := NewEmail("john.doe2@gmail.com")

	if err := contact.AddEmail(email1); err != nil {
		t.Error(err)
	}

	if err := contact.AddEmail(email2); err != nil {
		t.Error(err)
	}

	if err := contact.RemoveEmail(email1); err != nil {
		t.Error(err)
	}

	if emails := contact.GetEmails(); len(emails) != 1 {
		t.Error("expected 1 email to remain")
	}
}
