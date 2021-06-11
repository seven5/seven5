package seven5

//Authorizer is a type that allows implementors to control authorization and thus circumvent the usual
//rest dispatch machinery.  This type is consumed by RawDispatcher and should only be implemented by
//other dispatchers.  Applications should typically use the "Allow*" methods on their own resource
//implementation in combinations with the BaseDispatcher.
type Authorizer interface {
	Index(d *restShared, bundle PBundle) bool
	Post(d *restShared, bundle PBundle) bool
	Find(d *restObj, num int64, bundle PBundle) bool
	FindUdid(d *restObjUdid, id string, bundle PBundle) bool
	Put(d *restObj, num int64, bundle PBundle) bool
	PutUdid(d *restObjUdid, id string, bundle PBundle) bool
	Delete(d *restObj, num int64, bundle PBundle) bool
	DeleteUdid(d *restObjUdid, id string, bundle PBundle) bool
}

//AllowReader is an interface that allows a particular resource to express permissions about what users
//or types of requests are allowed on it.  This is a good place to put gross-level kinds of "policy"
//decisions like "non staff members cannot call this method".  This does not allow for very fine-grain policies about the _content_
//of the returned values, such as "the list returned has all elements if the user is a staff
//member but only elements they own for normal users."  AllowReader is used for the RestIndex interface.
//Return true to allow calls to RestIndex to be made with the PBundle provided.
type AllowReader interface {
	AllowRead(PBundle) bool
}

//AllowWriter is an interface that allows a particular resource to express permissions about what users
//or types of requests are allowed on it.  This is a good place to put gross-level kinds of "policy"
//decisions like "only staff members can create new instances of this resource".
//AllowWriter is used for the RestPost interface.  Return true to allow calls to RestPost to be made
//with the PBundle provided.
type AllowWriter interface {
	AllowWrite(PBundle) bool
}

//Allower is an interface that allows a particular resource to express permissions about what users
//or types of requests are allowed on it.  This is a good place to put gross-level kinds of "policy"
//decisions like "users may only write to their to objects they own". Allower is used for
//RestFind, RestPut, or rest delete.  The first parameter is the id of the resource.
//The second is the method of the request as as a string in uppercase, and the third is the parameter
//bundle that will be sent to the implementing method, if this method returns true.
type Allower interface {
	Allow(int64, string, PBundle) bool
}

//AllowerUdid is an interface that allows a UDID resource to express permissions about what users
//or types of requests are allowed on it.  This is a good place to put gross-level kinds of "policy"
//decisions like "users may only write to their to objects they own". Allower is used for
//RestFind, RestPut, or Rest Delete.  The first parameter is the id of the resource.
//The second is the method of the request as as a string in uppercase, and the third is the parameter
//bundle that will be sent to the implementing method, if this method returns true.
type AllowerUdid interface {
	Allow(string, string, PBundle) bool
}
