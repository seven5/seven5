package gauth

import (
	"github.com/seven5/seven5" 
)
//This is wire type that is accessible only to staff members.  It can only be read.
type GauthUserMetadata struct {
	Id seven5.Id
	NumberUsers seven5.Integer
	NumberGoogle seven5.Integer
	NumberLocal seven5.Integer
	NumberStaff seven5.Integer
}

//stateless
type GauthUserMetadataResource struct {
}

//Index can just return the metadata because the AllowRead function has already been called to check
//to see if it is ok for the logged in user to read this data.
func (STATELESS *GauthUserMetadataResource) Index(headers map[string]string,
	qp map[string]string, session seven5.Session) (string, *seven5.Error) {
	
	metadata := &GauthUserMetadata {
		seven5.Id(0),
		seven5.Integer(len(knownUsers)),
		seven5.Integer(len(knownUsers)),
		seven5.Integer(0),
		seven5.Integer(1),
	}
	list:=[]*GauthUserMetadata{metadata}
	return seven5.JsonResult(&list,true)
}

//used to create dynamic documentation/api
func (STATELESS *GauthUserMetadataResource) IndexDoc() *seven5.BaseDocSet {
	return &seven5.BaseDocSet{
		Headers: "This resource consumes no special headers.",

		Result: "This resource returns a collection of size one with one stance of GauthMetadata inside."+
		" This resource can only be successfully invoked by staff users.",

		QueryParameters: "This resource ignores all query parameters.",
	}
}




//AllowRead checks to insure that you have a session and you are staff before you can call
//this method.  This is the indexer and only method on this resource.
func (STATELESS *GauthUserMetadata) AllowRead(session seven5.Session) bool {
	u := GauthUserFromSession(session, false)
	//not logged in?
	if u == nil {
		return false
	}
	return u.isStaff 
}
