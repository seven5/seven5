package seven5

import (
	"testing"
)

func TestSessionBasic(t *testing.T) {
	mgr := NewSimpleSessionManager()

	s, err := mgr.Find("bogus")
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}
	if s != nil {
		t.Errorf("Unexpected find of 'bogus'")
	}
	
	s, err = mgr.Generate()
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}
	
	f, err := mgr.Find(s.SessionId())
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}
	
	if f.SessionId()!=s.SessionId() {
		t.Errorf("Unexpected session returned (expected %s but got %s)",s.SessionId(),f.SessionId())
		
	}
	
	err = mgr.Destroy("bogus")
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}

	err = mgr.Destroy(s.SessionId())
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}
	
	f, err = mgr.Find(s.SessionId())
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}
	
	if f!=nil {
		t.Errorf("Unexpected find of '%s'", s.SessionId())
	}
	
	if packetsProcessed != 6 {
		t.Errorf("Expected to have processed %d packets, but found %d\n", 6, packetsProcessed)
	}
	
}
