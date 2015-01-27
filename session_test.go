package seven5

import (
	"os"
	"strings"
	"testing"
	"time"
)

type testGen struct {
}

func (t *testGen) Generate(uniqueId string) (interface{}, error) {
	return uniqueId, nil
}

func TestSessionBasic(t *testing.T) {
	os.Setenv("SERVER_SESSION_KEY", strings.Repeat("0", 32))
	mgr := NewSimpleSessionManager(&testGen{})
	packetsProcessed = 0

	sr, err := mgr.Find("bogus")
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}
	if sr != nil {
		t.Errorf("Unexpected find of 'bogus'")
	}

	//create a new session
	ud, err := mgr.Generate("blah")
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}
	s, err := mgr.Assign("blah", ud, time.Time{})
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}
	if s.UserData().(string) != "blah" { //user data and unique id are same
		t.Errorf("Unexected user data, expected %s but got '%s'", "blah", s.UserData())
	}

	f, err := mgr.Find(s.SessionId())
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}
	if f == nil || f.Session == nil {
		t.Fatalf("Failed to find the session: %s", err)
	}

	if f.Session.SessionId() != s.SessionId() {
		t.Errorf("Unexpected session returned (expected %s but got %s)", s.SessionId(), f.Session.SessionId())
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

	if f != nil && f.Session != nil {
		t.Errorf("Unexpected find of '%s'", s.SessionId())
	}

	if f.UniqueId != "blah" {
		t.Errorf("did not recover unique id %s, got '%s'", "blah", f.UniqueId)
	}

	f, err = mgr.Find("garbagex")
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}
	if f != nil {
		t.Error("Failed to refuse bad session id: %s or %+v", f.UniqueId, f.Session)
	}
	expires := time.Now().Add(1 * time.Second)
	s, err = mgr.Assign("fleazil", ud, expires)
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}
	sr, err = mgr.Find(s.SessionId())
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}
	if sr.Session == nil {
		t.Errorf("didn't look up the session properly: %+v", sr)
	}
	time.Sleep(1 * time.Second) //at least 1 sec
	sr, err = mgr.Find(s.SessionId())
	if err != nil {
		t.Fatalf("Failed to communicate to the session manager: %s", err)
	}
	if sr != nil {
		t.Errorf("session expired, should not recover anything: %+v ", sr)
	}

	//
	// Test to make sure we sent the right number of packets through the
	// channel
	//

	if packetsProcessed != 10 {
		t.Errorf("Expected to have processed %d packets, but found %d\n", 6, packetsProcessed)
	}

}
