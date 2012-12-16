package seven5

import (
	"code.google.com/p/goauth2/oauth"
	"fmt"
)

const (
	GOOGLE_AUTH_URL      = "https://accounts.google.com/o/oauth2/auth"
	GOOGLE_TOKEN_URL     = "https://accounts.google.com/o/oauth2/token"
)

//GoogleAuthConn is the implementation of an AuthServiceConnector for Google.
type GoogleAuthConn struct {
	Scope  string
	Prompt string
}

//NewGoogleAuthConnector returns an AuthServiceConnector suitable for use with Google.  Note
//that the scope variable is very important to google and should be set based on the needs
//of your app.  The prompt values can be "auto" or "force" to force a re-prompt from google
//on each authentication handshake.
func NewGoogleAuthConnector(scope string, prompt string, info AppAuthConfig) AuthServiceConnector {
	result := &GoogleAuthConn{Scope: scope, Prompt: prompt} //, Info: info}
	return result
}

//OauthConfig creates the config structure needed by the code.google.com/p/goauth2 library.
func (self *GoogleAuthConn) OauthConfig(info AppAuthConfig) *oauth.Config {
		
	return &oauth.Config{
		ClientId:       info.ClientId(self),
		ClientSecret:   info.ClientSecret(self),
		Scope:          self.Scope,
		AuthURL:        GOOGLE_AUTH_URL,
		TokenURL:       GOOGLE_TOKEN_URL,
		RedirectURL:    fmt.Sprintf("%s%s",info.RedirectHost(self),info.RedirectPath(self)),
		ApprovalPrompt: self.Prompt,
	}
}

//ExchangeForToken returns an oauth.Token structure from a code received in the handshake
//plus the basic information in the AppAuthConfig.  This will be called after the
//first phase of the oauth exchange is done and we want to exchange a code for a token.
func (self *GoogleAuthConn) ExchangeForToken(info AppAuthConfig, code string) (*oauth.Token, error) {
	config := self.OauthConfig(info)
	//exchange it
	t := &oauth.Transport{Config: config}
	tok, err := t.Exchange(code)
	if err != nil {
		return nil, err
	}
	return &oauth.Token{AccessToken: tok.AccessToken}, nil
}

func (self *GoogleAuthConn) Name() string {
	return "google"
}

func (self *GoogleAuthConn) AuthURL(info AppAuthConfig, state string) string {
	return self.OauthConfig(info).AuthCodeURL(state)
}

func (self *GoogleAuthConn) CodeValueName() string {
	return "code"
}
func (self *GoogleAuthConn) ErrorValueName() string {
	return "error"
}

