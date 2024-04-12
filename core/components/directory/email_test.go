package directory

import "testing"

func TestEmail_Valid(t *testing.T) {

	if _, err := NewEmail("john.doe@gmail.com"); err != nil {
		t.Error(err)
	}
}

func TestEmail_Valid2(t *testing.T) {
	
	if _, err := NewEmail("foo"); err == nil {
		t.Error("expected bad email type")
	}
}
