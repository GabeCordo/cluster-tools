package in_memory

import (
	"github.com/FortifiedCode/flock/internal/core/database/contact"
	"testing"
)

func Test_StatisticLocalDatabase_Is_Implemented(t *testing.T) {

	d := New()

	var i contact.Database
	i = d

	i.Print()
}

func TestDirectory_AddContact(t *testing.T) {

	directory := New()

	c := contact.NewContact("foo")
	if err := directory.AddContact(c); err != nil {
		t.Error(err)
	}

	if err := directory.AddContact(c); err == nil {
		t.Error("expected rejection of duplicate contact")
	}
}

func TestDirectory_DeleteContact(t *testing.T) {

	directory := New()

	c := contact.NewContact("foo")
	if err := directory.AddContact(c); err != nil {
		t.Error(err)
	}

	c2 := contact.NewContact("foo2")
	if err := directory.AddContact(c2); err != nil {
		t.Error(err)
	}

	if err := directory.DeleteContact("foo"); err != nil {
		t.Error(err)
	}

	if contacts := directory.GetContacts(); len(contacts) != 1 {
		t.Error("expected 1 contact to remain")
	}
}
