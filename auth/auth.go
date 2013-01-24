package auth

import (
	"code.google.com/p/goauth2/oauth"
)

//ServiceConnector is an abstraction of a service that can do Oauth-based authentication.
//For now this is a very thin wrapper over code.google.com/p/goauth2/oauth.
type ServiceConnector interface {
	AuthURL(string, string) string
	CodeValueName() string
	ErrorValueName() string
	StateValueName() string
	Name() string
	Fetch(*oauth.Transport) (interface{},error)
	ExchangeForToken(string, string) (*oauth.Transport, error)
}


//OauthClientDetail is an interface for finding the specific information needed to connect to
//an Oauth server.  If you don't want to use environment variables as the way you store
//these, you can provide your own implementation of this class.  Note that it can be called
//with different AuthServiceConnectors if you have multiple oauth providers.
type OauthClientDetail interface {
	ClientId(ServiceConnector) string
	ClientSecret(ServiceConnector) string
}
