package seven5

//Id is unique identifier for objects.  Ids are never negative but may be zero.  It
//is at least 64 bits wide.  Any struct that has this type as a member must have that
//member named "Id" and must be a resource.
type Id int64
