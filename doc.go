//An opinionated, stiff web framework written in go and and javascript.
//
//For general information about the project see http://seven5.github.com/seven5
//Volga Announcement: https://github.com/seven5/seven5/wiki/ANN:-Seven5-Release-Volga-(0.1):-First-Public-Release
//
//For information about the API, see the http://gopkgdoc.appspot.com/pkg/github.com/seven5/seven5 and  
//http://gopkgdoc.appspot.com/pkg/github.com/seven5/seven5/store, thoughtfully provided by godoc.
//
//Seven5 is very opinionated.   It's opinionated about *both* the client and the server 
//side--because these days you need them to work in concert to construct any serious app.
//
//Some opinions of Seven5:
//
//* Server side opinions:
//
//** Write server-side code in Go
//** No relational database by default
//** Assume key-value store 
//** No templating
//** No user-visible URL mapping (only APIs are exposed, not dynamic "content")
//** Expose objects as Go structs--exported to client side as "models" for Backbone
//** Expose services as REST 
//** Easy deployment to the cloud
//** Develop and test on a single machine
//
//* Client Side Opinions:
//
//** All presentation code on client-side 
//** Program in Javascript using civilized tools
//** MVC on client-side with backbone.js
//** Full support for test-driven development
//** All user-visible URLs controlled by client code (backbone's router)
package seven5

