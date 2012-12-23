package seven5


//AllowReader is an interface that allows a particular resource to express permissions about what users
//or types of requests are allowed on it.  This is a good place to put gross-level kinds of "policy"
//decisions like "non staff members cannot call this method".  This does not allow for very fine-grain policies about the _content_
//of the returned values, such as "the list returned has all elements if the user is a staff
//member but only elements they own for normal users."  AllowReader is used for Indexers.
//The parameter, if not nil, is the currently logged in browser's session.
type AllowReader interface {
	AllowRead(Session) bool
}

//AllowWriter is an interface that allows a particular resource to express permissions about what users
//or types of requests are allowed on it.  This is a good place to put gross-level kinds of "policy"
//decisions like "only staff members can create new instances of this resource".  
//AllowWriter is used for Posters.
//The parameter, if not nil, is the currently logged in browser's session.
type AllowWriter interface {
	AllowWrite(Session) bool
}

//Allower is an interface that allows a particular resource to express permissions about what users
//or types of requests are allowed on it.  This is a good place to put gross-level kinds of "policy"
//decisions like "users may only write to their to objects they own". Allower is used for 
//Finders, Puters, and Deleters.
//The first parameter is the id of the resource.  The second is the method of the request as
//as a string in uppercase, and the third, if not
//nil, is the currently logged in browser.
type Allower interface {
	Allow(Id,string,Session) bool	
}
