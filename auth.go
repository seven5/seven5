package seven5

import (
	//oauth1 "github.com/iansmith/go-oauth/oauth"
	oauth1 "github.com/garyburd/go-oauth/oauth"
	"net/http"
)

type OauthCred interface {
	Token() string
	Secret() string
}

type SimpleOauthCred struct {
	tok *oauth1.Credentials
}

func (self *SimpleOauthCred) Token() string {
	return self.tok.Token
}

func (self *SimpleOauthCred) Secret() string {
	return self.tok.Secret
}

//OauthConnection is a per-client connection to the service that the connection was
//returned from.  Its only method is the same as "http.Client.Do()" but with specific
//authentication machinery for the connector.  Note that there are failures of the
//authentication machinery and http failures lumped together on the error return value.
type OauthConnection interface {
	//Note that we don't want to expose an http.Client because oauth1 wants to do this
	//in a simple way of adding a single header to the output.
	SendAuthenticated(*http.Request) (*http.Response, error)
}

//OauthConnector is an abstraction of a service that can do Oauth-based authentication.
//This abstraction papers over, badly, the differences between oauth2 and oauth1(a).
type OauthConnector interface {
	ClientTokenValueName() string
	CodeValueName() string
	ErrorValueName() string
	StateValueName() string
	//Phase is not used by Oauth2 so it returns nil for OauthCred
	Phase1(state string, callbackPath string) (OauthCred, error)
	//Both versions use the state passed in here to help you know what to do when
	//you land on the login page
	UserInteractionURL(p1creds OauthCred, state string, callbackPath string) string
	//Oauth2 client token?
	Phase2(clientToken string, code string) (OauthConnection, error)
	Name() string
}

//OauthClientDetail is an interface for finding the specific information needed to connect to
//an Oauth server.  If you don't want to use environment variables as the way you store
//these, you can provide your own implementation of this class.
type OauthClientDetail interface {
	ClientId(serviceName string) string
	ClientSecret(serviceName string) string
}
