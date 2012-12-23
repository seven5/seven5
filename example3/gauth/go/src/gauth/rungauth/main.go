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
	LOGIN_TYPE = "auto" //can be "force" or "auto"
)

func main() {
	
	//supply our app level session manager and bind our app name (so we have cookies under our name)
	cm := seven5.NewSimpleCookieMapper(NAME, gauth.NewSessionManager())

	//create a new, empty URL space... needs the cookie mapper because it will try to throw
	//cookies away if not valid (seen during dispatching of URLs)
	h := seven5.NewSimpleHandler(cm)

	//add a user WIRE type to URL space, not our internal gauth.GauthUserInternal
  h.AddResource(gauth.GauthUser{}, &gauth.GauthUserResource{})

	//we use environment variables to get at things not in the source (and so does heroku)... 
	//NAME is prefixed to everything
	env := seven5.NewEnvironmentVars(NAME)
	//this is used to help us know our hostname in production (when testing, uses localhost)
	deploy := seven5.NewRemoteDeployment(seven5.HerokuName(NAME), env)

	//authentication via google, it needs to know some secrets from the env and the 
	//hostname for production
	goog := seven5.NewGoogleAuthConnector(SCOPE, LOGIN_TYPE, env, deploy)
	//where we are going to go based on various interactions with google authentication
	pm := seven5.NewSimpleAuthPageMapper("error.html", "home.html", "home.html", deploy)
	
	//bind all the oauth stuff together and stuff into URL space
	seven5.AddAuthService(h, pm, goog, cm)

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
