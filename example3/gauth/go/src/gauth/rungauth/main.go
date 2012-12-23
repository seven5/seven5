package main

import (
	"errors"
	"fmt"
	"net/http"
	"gauth"
	"seven5"//githubme:seven5:
)

var BAD_WEB_ROOT = errors.New("Cant access web root")

const (
	NAME = "gauth"
	SCOPE = "https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email "
	FORCE = "auto" //can be "force" or "auto"
)

func main() {
	
	//supply our session generator function and bind our app name (so we have cookies under our name)
	sess := seven5.NewSimpleCookieMapper(NAME, gauth.NewSessionManager())

	//create a new, empty URL space
	h := seven5.NewSimpleHandler(sess)

	//add a user WIRE type to URL space, not our internal gauth.GauthUserInternal
  h.AddResource(gauth.GauthUser{}, &gauth.GauthUserResource{})

	//create google login connector plus heroku deployment info (and also testing on localhost)
	cfg := seven5.NewHerokuAppAuthConfig(NAME)
	goog := seven5.NewGoogleAuthConnector(SCOPE, FORCE)
	
	//bind all the oauth stuff together and stuff into URL space
	seven5.AddAuthService(h, cfg, goog, sess)

	//normal seven5 setup for project put in URL space... note that this has to be AFTER you added
	//resources because it causes dart code generation for those resources
	seven5.DefaultProjectBindings(h, NAME)

	//run
	err := http.ListenAndServe(":3003", logHTTP(h))

	fmt.Printf("Error! Returned from ListenAndServe(): %s", err)
}

func logHTTP(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
