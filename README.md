# Seven5

## An opinionated, stiff web framework written in go.

For general information about the project see [seven5's documentation on github](http://seven5.github.com/seven5).

For information about the API, see the [seven5 package doc](http://gopkgdoc.appspot.com/pkg/github.com/seven5/seven5) and the [store package doc](http://gopkgdoc.appspot.com/pkg/github.com/seven5/seven5/store), thoughtfully
provided by godoc.

Seven5 is very opinionated.   It's opinionated about *both* the client and the server side--because these days you need them to work in concert to construct any serious app.

## Some opinions of Seven5

### Server side opinions:

* Write in Go
* No relational database by default
* Assume key-value store 
* No templating
* No user-visible URL mapping (only APIs are exposed, not dynamic "content")
* Expose objects as Go structs--exported to client side as "models" for Backbone
* Expose services as REST 
* Easy deployment to the cloud
* Develop and test on a single machine

### Client Side Opinions:

* Program in Javascript using civilized tools
* MVC on client-side with backbone.js
* Full support for test-driven development
* All presentation code in client code
* All user-visible URLs controlled by client code (backbone's router)

