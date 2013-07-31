---
layout: book
chapter: Understanding Resources
---

### Goal: WireTypes and Resources
At the end of this chapter, you should understand the `ArticleWire` and `ArticleResource` types that in the source code for this chapter.  You should also understand the role of a _wire type_ and _resource implementation_ in a _Seven5_ application based on REST.   For this chapter, you can continue to use the code you checked out two chapters ago, based on the branch "code-book-1". 

### Theory: Sloganeering
You have, by now, noticed the slogan, "Restful without remorse or pity" on various _Seven5_ web pages.  This is intended to convey that _Seven5_ supports only REST-based back ends for any web applications.  This is true, but perhaps a bit misleading, as nothing prevents you from using Go's existing web server development tools to develop the non-REST portions of the application.

### Theory: Designing for REST and Seven5
REST is far from complete in terms of coverage of all the ideas one needs to build a working web application. That said, it is also not too bad, so this book will just bend and twist when necessary to make REST work.  To design a new application based on REST principles, one should start with the _nouns_ that are key to the system and make sure that these nouns make sense with each of the following verb phrases.  It's ok if "makes sense" boils down to "that is not possible" but it should be quite clear why "not possible" is a sensible answer.  A REST interface to our solar system may not allow the creation of new planets.

* _Get index_ (GET) - Get a list of all the instances of the noun
* _Create for effect_ (POST) - Make a new instance of the noun based on the properties provided
* _Get the properties_ (GET) - Get the properties of an existing noun, by Id
* _Delete for effect_ (DELETE) - Destroy an existing noun, by Id
* _Update all the properties_  (PUT) Change the properties of the existing noun, by Id, supplying a complete set of new properties
* _Update a single property_ (PATCH) Change a single property of the existing noun, by Id, supplying a single new property value and leaving the others unchanged

When implementing REST over HTTP, as in the case of _Seven5_, it is customary to use the HTTP "verbs" that are shown in parenthesis as a shorthand for these operations.  All the verbs typically return the complete set of properties for the object being created or operated on, except the index verb.  When referring to an object in the URL space, it is customary to name the object by an id after the noun  _in the singular and all lower case_.  Thus

```
GET /car/234
```

Indicates an HTTP request ("GET"), usually but not always from a browser, seeking to get properties of the car (noun object) with Id 234.  In the interest of simplicity, we will follow these conventions throughout this book, although we prefix all references to the REST namespace with `/rest` to prevent confusion with applications that also need a non-REST portion.

It should be noted above that "GET" does double duty depending on if it is called as `/rest/car/` (show all cars) or `/rest/car/123` (show properties of car with Id 123).  "PATCH" is not yet supported by all browsers, so it may be simulated with some client-side code and the use of "PUT".  "PATCH" was added after the initial five other verbs were created, to ease the burden of constantly being forced to send the entire, possibly large, set of properties to the server in order to change a single property. 

### Theory: The wire type and APIs

REST defines a set of verbs to apply to an application-specific set of nouns.  As we have seen, the network protocol (HTTP over TCP) and the definition of the URL space have well-understood customs as well.  What's left to do in terms of writing the "API" of an application?  The names of the nouns that are understood by the application, and their properties. 

The _wire type_ associated with a noun in _Seven5_ defines this API by exposing a set of fields that will be exchanged between the client and server via REST.  It is important to note that just because a particular API is defined, this says little or even nothing about the storage model used _inside_ the server implementation. It is common for a server implementing a noun like "animal" to have a set of fields like "number_legs" and "vegetarian" that are exchanged with the client, some of which end up being stored on the server, some of which are perhaps used internally by the server but not stored, and other fields that are stored for the exclusive use of the server and not shared with the client. The wire type defines defines what passes over the wire, nothing more.

### Practice: The Wire Type

This book details many conventions, or perhaps better, "expectations", about the way _Seven5_ development works in practice. For wire types, it is assumed that the name will be `FooWire` for an application-specific noun "foo".  Because of Go's convention of exposing types to external packages if the name is capitalized,   `FooWire` is publicly visible throughout the application.

Returning to _nullblog_, one cannot have a blog without articles, so we will start there.  This source code can be found in `/tmp/book/go/src/nullblog.go`

```go

type ArticleWire struct {
        Id      seven5.Id
        Content seven5.String255
        Author  seven5.String255
}
```

There are three key constraints on wire types:

* Wire types must contain an Id field with the type `seven5.Id`
* Any other fields in a wire type must have a type that comes from the package `seven5` 
* Field names must be capitalized  (implied: there is no point in a private field in a wire type)

A full list of the types known to _Seven5_ and permitted in a wire type can be found in the file `types.go` in the _Seven5_ source.  A future chapter will explain more about the options you have here.

The type used here for `Content` and `Author` are "short" string fields, less than 256 characters. This clearly is a silly requirement for the content of a blog post, but it will corrected in a later chapter. _Seven5_ will automatically generate the JSON values associated with sending `ArticleWire` to the client, and reversing that process when receiving `ArticleWire`, but can only do this with fields that are visible to it.  Thus, Go's rule about capitalization comes into play, and all fields in a wire type must be capitalized.

### Practice: Defining the implementation, the stateless "resource"

The implementation of a noun such as article is done with a type called `ArticleResource`.  Again, from the file `/tmp/book/go/src/nullblog.go`

```go
type ArticleResource struct {
        //STATELESS
}
```

It is frowned upon by both the REST originator ([Roy Fielding](http://en.wikipedia.org/wiki/REST)) and this author to hold state in a resource. The expectation for a type like `ArticleResource` is that it contains _exclusively_ methods, and that a method handles the implementation of a particular verb.  The input to a method is typically either an Id, specifying which article is to be affected, or an instance of `ArticleWire` indicating the content of an article object to be updated.

>>>> When one is concerned about scalability, having resources be stateless is  a significant boon.  Typically, adding more "web tier" nodes can be done quite easily when this statelessness constraint is observed.

The code for this chapter has only two of the REST methods implemented.  Both of these implementations return simple, fixed responses to `GET /rest/article`, `GET /rest/article/0` and `GET /rest/article/1`. 

```go
var someArticle = []*ArticleWire{
        &ArticleWire{
                0, "This is a really short article that demonstrates the concept.", "Ian Smith",
        },
        &ArticleWire{
                1, "Another very short article, must be less than 255 characters!", "Ian Smith",
        },
}

func (IGNORED *ArticleResource) Index(bundle seven5.PBundle) (interface{}, error) {
        return someArticle, nil
}

func (IGNORED *ArticleResource) Find(Id seven5.Id, bundle seven5.PBundle) (interface{}, error) {
        i := int64(Id)
        if i < 0 || int(i) > len(someArticle) {
                return nil, seven5.HTTPError(http.StatusBadRequest, "nice try, loser")
        }
        return someArticle[i],nil
}
``` 

Worth noting in these two implementation methods:

* The return value of a non-error case in a resource implementation is an instance of the wire type or, in the case of Index, a slice of zero or more wire types.   _Seven5_ handles all the necessary marshaling to send JSON back to the client.

* The error cases simply return an `error` in the second return value with nil as the first value, as this is the [standard](http://blog.golang.org/error-handling-and-go) way of handling errors in Go.  There are some convenience methods in _Seven5_, such as `HTTPError` used in the `Find` method to send particular HTTP status codes back to the client.  If you return a Go standard error object, a 500 "Internal Error" status code is sent to the client.

* Implementations of resources _must not_ trust that the parameters sent by clients are trustworthy.  In the method `Find` above the implementation checks that the Id value provided as the first parameter is in bounds for the set of fixed objects that this implementation knows about.

>>>> The strategy used above is a common one during _Seven5_ development.  This simple, "fixed set of data" implementation can be used to allow client-side development to proceed without needing a full implementation of the server part of the API.

### Practice: The PBundle

The type `seven5.PBundle` represents a parameter bundle; this bundle is data collected from the HTTP request that caused a resource method to be called.  In the cases above, the `PBundle` is ignored but it can be used to access "out of band" information from the HTTP request. The `PBundle` gives access to query parameters from the particular URL invoked, the request headers, and session information such as the currently logged in user. These will more fully detailed in later chapters.  The use of the `PBundle` to access form parameters is strongly discouraged as this is better handled by other mechanisms.

### Practice: The REST interfaces

```go
type RestIndex interface {
        Index(PBundle) (interface{}, error)
}
type RestFind interface {
        Find(Id, PBundle) (interface{}, error)
}
type RestDelete interface {
        Delete(Id, PBundle) (interface{}, error)
}
type RestPut interface {
        Put(Id, interface{}, PBundle) (interface{}, error)
}
type RestPost interface {
        Post(interface{}, PBundle) (interface{}, error)
}

```

Above are the types in the library `seven5` that define the interface to the verb phrases for REST services.  To implement a particular verb, you should implement the appropriate method from the list above in the resource implementation, not the wire type.  Slogan: Wire types are data only, resources hold no data and consume and return wire types.

Each type name is "Rest" plus the name of the HTTP operation, except that `seven5.RestFind` and `seven5.RestIndex` are used to differentiate the two types of GET messages. Since a given resource implementation type may choose which of the interfaces above to implement, some of the methods may not be available on a particular resource.  Attempts to use such methods result in an HTTP result code of 501, with the message "Not implemented".  For example, in the code above for `ArticleResource` only the Index and Find methods are implemented, so attempts to use other REST methods on an article will fail--in the sense that they will not return an HTTP status code in the 200's.


