package directory

import "testing"

func TestDirectory_AddContact(t *testing.T) {

	directory := NewDirectory()

	contact := NewContact("foo")
	if err := directory.AddContact(contact); err != nil {
		t.Error(err)
	}

	if err := directory.AddContact(contact); err == nil {
		t.Error("expected rejection of duplicate contact")
	}
}

func TestDirectory_DeleteContact(t *testing.T) {

	directory := NewDirectory()

	contact := NewContact("foo")
	if err := directory.AddContact(contact); err != nil {
		t.Error(err)
	}

	contact2 := NewContact("foo2")
	if err := directory.AddContact(contact2); err != nil {
		t.Error(err)
	}

	if err := directory.DeleteContact("foo"); err != nil {
		t.Error(err)
	}

	if contacts := directory.GetContacts(); len(contacts) != 1 {
		t.Error("expected 1 contact to remain")
	}
}
