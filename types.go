package seven5

import (
	"strings"
)

//Floating is _at least_ a 64bit IEEE754 buzzword-compliant double-precision number.
type Floating float64

//String255 is a string that can be assumed to fit in a fixed size, 255 byte buffer if
//it needs to be sent to disk.  Since the encoding may vary, this does not imply
//that it can hold 255 charactere.
type String255 string

//Textblob is a large chunk of text.  This size of text object cannot be assumed to fit
//in a fixed size buffer.  A Textblob is suitable for an SQL "text" field and thus may be
//expensive to search for.
type Textblob string

//Integer is at least a 64 bit integer.  
type Integer int64

//Id is unique identifier for objects.  Ids are never negative but may be zero.  It
//is at least 64 bits wide.  Any struct that has this type as a member must have that
//member named "Id" and must be a resource.
type Id int64

//Boolean is either true or false. 
type Boolean bool

func TrimSpace(s String255) String255 {
	return String255(strings.TrimSpace(string(s)))
}
