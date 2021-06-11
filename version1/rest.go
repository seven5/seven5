package seven5

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

type RestIndex interface {
	Index(PBundle) (interface{}, error)
}
type RestFind interface {
	Find(int64, PBundle) (interface{}, error)
}
type RestFindUdid interface {
	Find(string, PBundle) (interface{}, error)
}
type RestDelete interface {
	Delete(int64, PBundle) (interface{}, error)
}
type RestDeleteUdid interface {
	Delete(string, PBundle) (interface{}, error)
}
type RestPut interface {
	Put(int64, interface{}, PBundle) (interface{}, error)
}
type RestPutUdid interface {
	Put(string, interface{}, PBundle) (interface{}, error)
}

type RestPost interface {
	Post(interface{}, PBundle) (interface{}, error)
}

type RestAll interface {
	RestIndex
	RestFind
	RestDelete
	RestPost
	RestPut
}

type RestAllUdid interface {
	RestIndex
	RestFindUdid
	RestDeleteUdid
	RestPost
	RestPutUdid
}

type restShared struct {
	typ   reflect.Type
	name  string
	index RestIndex
	post  RestPost
}

type restObj struct {
	restShared
	find RestFind
	del  RestDelete
	put  RestPut
}

type restObjUdid struct {
	restShared
	find RestFindUdid
	del  RestDeleteUdid
	put  RestPutUdid
}

//
// IsUDID takes in a string and returns true if it is formatted as a standard
// UDID, for example de305d54-75b4-431b-adb2-eb6b9e546013.  This code expects
// the length to be exactly 36 characters, with five groups of hex characters
// separated by dashes.  The five groups are of length 8, 4, 4, 4, and 12
// characters.

func IsUDID(s string) bool {
	if len(s) != 36 {
		return false
	}
	parts := strings.Split(s, "-")
	if len(parts) != 5 {
		return false
	}
	for k, v := range map[int]int{0: 8, 1: 4, 2: 4, 3: 4, 4: 12} {
		if len(parts[k]) != v {
			return false
		}
	}
	for _, r := range s {
		switch r {
		case 'a', 'b', 'c', 'd', 'e', 'f', '-', 'A', 'B', 'C', 'D', 'E', 'F', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			return false
		}
	}
	return true
}

//UDID returns a new UDID, formatted such that it wil pass IsUDID(), by
//reading /dev/urandom.  This value contains 16 bytes of randomness.
func UDID() string {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		panic(fmt.Sprintf("failed to get /dev/urandom! %s", err))
	}
	b := make([]byte, 16)
	_, err = f.Read(b)
	if err != nil {
		panic(fmt.Sprintf("failed to read  16 bytes from /dev/urandom! %s", err))
	}
	f.Close()
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
