package seven5

//Id is unique identifier for objects.  Ids are never negative but may be zero.  It
//is at least 64 bits wide.  Any struct that has this type as a member must have that
//member named "Id" and must be a resource.
type Id int64

//String255 is a sequence of characters that has a hard limit on its
//length of 255.
type String255 string

//Textblob is a sequence of characters of unlimited size
type Textblob string

//DateTime is a representation of a moment in time.  This has no
//timezone, thus is always UTC.  The maximum precision here is
//seconds.
type DateTime int
