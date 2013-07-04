package main

import (
        "fmt"
        "myproj" //your library
        "net/http"
        "github.com/seven5/seven5" //ungithubme:seven5:
)

const (
        NAME = "myproj"
        REST = "/rest/"
)

func main() {
        //we use the environment variables and the heroku name for our app deployment URL.
        //Environment var PORT controls the port number we run on, both for production and test.
        //Set GITVET_TEST non empty for use on localhost.
        heroku := seven5.NewHerokuDeploy(NAME)
        mux := seven5.DefaultProjectBindings(NAME, heroku.Environment(), heroku)

        //a dispatcher takes in raw requests and picks the appropriate API to dispatch them on
        //base dispatcher works with rest resources and understands about the "Allow" interfaces
        bd := seven5.NewBaseDispatcher(NAME, nil)

        // the default location is /rest for the resources inside bd
        mux.Dispatch(REST, bd)
        //implementation resources
        sub:=&myproj.GreetingResource{}
        bd.ResourceSeparate("Greeting", &myproj.GreetingResource{}, sub, nil, 
                nil, nil, nil)

        http.ListenAndServe(fmt.Sprintf(":%d", heroku.Port()), mux)
