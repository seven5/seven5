---
layout: book
chapter: Understanding main
---

### Goal: Project Layout
After reading this chapter, you should have a basic familiarity with the `main()` function used in a _Seven5_ server-side application and some understanding of the options available to a developer based on the functions used in `main()`. For this chapter, you can continue to use the code you checked out in the first chapter, based on the branch "code-book-1". 

### Theory: Is less more in main()?
The author has considered trying to shorten `main()` functions for simple _Seven5_ applications like _nullblog_.  The idea would be to have a layer over the code explored in this lesson that allows a developer to ignore many of these choices. To this point, this "shorter main approach"" has been eschewed in favor of a pedagogical approach which can be explained as "copy this code and glance at it to see if you need to throw any of these switches."  It is far from clear that the approach taken to date is superior.

### Practice: main()
In the file `/tmp/book/go/src/nullblog/runnullblog/main.go` you can find the source code to the command `runnullblog` which includes the entry point of the program. At this entry point, several objects are created and configured that connect the application to the infrastructure of _Seven5_.  The `main()` function and a couple of constants, comments removed for space, is

```go

const (
        NAME = "nullblog"
        REST = "/rest/"
)

func main() {
        heroku := seven5.NewHerokuDeploy(NAME)
        mux := seven5.DefaultProjectBindings(NAME, heroku.Environment(), heroku)

        bd := seven5.NewBaseDispatcher(NAME, nil)
        mux.Dispatch(REST, bd)

        articleRez:=&nullblog.ArticleResource{}
        bd.ResourceSeparate("article", &nullblog.ArticleWire{}, articleRez, articleRez, 
                nil, nil, nil)

        http.ListenAndServe(fmt.Sprintf(":%d", heroku.Port()), mux)
}
```

### Practice: Naming within _Seven5_

_Seven5_ follows numerous conventions when it exposes its public API.  The most important ones are:

* Creating an object of a public type "Foo" is done with the function "NewFoo".  This is the idiomatic-Go way of naming creation functions.
* When _Seven5_ provides an implementation of a interface "Bar" that is intended to be good enough for most applications, that implementation is named "SimpleBar".  To create an instance of SimpleBar, as above, the function "NewSimpleBar" would be exposed by the API.
* When _Seven5_ provides a function that should be used by applications that don't intend to change the existing behavior of _Seven5_ itself, those names begin with "Default".  

### Practice: Deployment Objects

Above, the variable `heroku` is assigned an instance of type `seven5.HerokuDeployment *`.  How to _deploy_ an application to [Heroku](http://www.heroku.com) will be explored in a later chapter; already we have seen how to run an application locally.  Because these two tasks are intricately linked, we have already been using the `seven5.HerokuDeployment` object.  It is worth noting that the `PORT` environment variable, used in a previous chapter to control the port that our simple _nullblog_ server runs on, is identical to the way Heroku behaves in a production setting.

>>>> For those interested in supporting other deployment or test scenarios, the type `seven5.RemoteDeployment` is worth investigating in the source code.  This is almost certainly insufficient at the current time to support varying deployment targets, but is intended to be the place that common functionality for these tasks live.

### Practice: ServeMux Objects

In the `main()` above the variable `mux` is assigned the type `seven5.ServeMux *`, the return value of `seven5.DefaultProjectBindings`.  By virtue of duck typing, `mux` is also an instance of `http.ServeMux *`, familiar to most Go developers already from the "net/http" package.  

The choice to use `seven5.DefaultProjectBindings` is right for most applications because it configures the `ServeMux` to serve static content from the project's web directory (`/tmp/book/dart/nullblog/web/`) when requested from the URL `/`.  The value `heroku.Environment()` is passed into the function `DefaultProjectBindings` to allow access to the environment variable `GOROOT`, as explained in the previous chapter.  The function's name refers to "Project" bindings because this method depends crucially on the default project layout that was explained in a prior chapter.

### Theory: Configuring the URL space

The careful reader will note that a deployment object, in this case `heroku`, is passed to `DefaultProjectBindings`.  This is done to allow the deployment to manipulate the URL space to accommodate the needs of a particular deployment/test strategy.  Because the `mux` variable is a standard [ServeMux](http://golang.org/pkg/net/http/#ServeMux) objects, applications that want to do so may add URL bindings to it in the normal fashion.

 
### Practice: Dispatchers

The interface `seven5.Dispatcher` and related types will be covered in a future chapter. It is sufficient for now to understand that a dispatcher's job is to convert an HTTP request into a function call on a particular interface.  Implied in this is that a dispatcher must marshal and unmarshal wire types to convert function call arguments and return values to a format suitable for input or output over the network.  

Here are the two lines from `main()` that are relevant to dispatchers:

```go
bd := seven5.NewBaseDispatcher(NAME, nil)
mux.Dispatch(REST, bd)
```

This code creates an instance of `seven5.BaseDispatcher *` and that is assigned to the variable `bd`.  Then the `seven5.ServeMux *` contained in `mux` is used to bind the entire REST namespace, `/rest`, to the base dispatcher just created.  "BaseDispatcher" is so named because is intended to be used by the authors of other types who want to borrow chunks of its functionality.

### Theory: Designing for testing

For those readers that are [test infected](http://c2.com/cgi/wiki?TestInfected), a less obvious goal in the design of dispatchers can be seen.  As we will see in the next chapter, it is easy to test a server implementation such as `ArticleResource` because it is completely decoupled from the HTTP machinery.  It is worth examining the code in `/tmp/book/go/src/nullblog/nullblog.go` to discover that the _only_ use of the Go "net/http" package in that code is to access, and avoid repeating, the HTTP status codes.  Otherwise, _nullblog's_ server-side has no "network connection".

Dispatchers themselves perform only the translation from HTTP to the interface form, such as `seven5.RestIndex` or `seven5.RestFind` in our sample code to this point. Thus, they are also easier to test as they are quite specialized and small, although they do require testing with real HTTP messages and a real network. 

### Practice: Associating resources 
Returning to our discussion of `main()` in `runnullblog` one could argue the two most important tasks in the main of a server implementation are those that (1) associate the wire type with its implementing resource and (2) binding that combination to a portion of the URL space.  These are done like this:

```go
articleRez:=&nullblog.ArticleResource{}
bd.ResourceSeparate("article", &nullblog.ArticleWire{}, articleRez, articleRez, 
        nil, nil, nil)
```

The first line simply creates an instance of our resource type and assigns a pointer to it to `articleRez`. It  is customary in _Seven5_ development to use "Rez" as a suffix when one has a pointer to a resource implementation.  The second line does both of the tasks referred to above.  Each parameter has an important job in the two tasks:

* The function `ResourceSeparate` means that we want to bind each of the REST interface methods in the previous chapter, such as `seven5.RestFind` or `seven5.RestPost`, separately.  This allows us to pass `nil` for REST verbs that we do not wish to implement.  The final three `nil` parameters indicate that we do not wish to implement `seven5.RestPost`, `seven5.RestPut`, and `seven5.RestDelete`, in that order.
* The string "article" tells the BaseDispatcher `bd` the name of this rest noun.  It also allocates the "/rest/article" part of the URL space for our implementation.
* The code `&nullblog.ArticleWire{}` creates an _examplar_ of our wire type and associates it with our implementation methods that immediately follow.  (This is necessary because Go's reflection system is not sufficiently powerful to take the name of a type as a string and convert that to an instance of that type.  Thus, we must provide this exemplar via code.)
* The two mentions of `articleRez` indicate that this object is providing the implementation of `seven5.RestIndex` and `seven5.RestFind` associated with the exemplar's wire type. Note that one could provide these implementations with two instances rather than the same instance as is done here, but there is no need to do so as `articleRez` is stateless and only its code is used.

### Theory: interface{} or ArticleWire*

The careful reader will note that if you make gross mistake in the type signature of your REST implementation methods such `seven5.RestFind` it will be detected at compile-time. The type signatures are checked when compiled at the point of the call to `ResourceSeparate()` above.  However, the wire type could be in error and this will not be detected until run-time, when an abort will be forced.
 
The POST and PUT methods take a parameter of type `interface{}` in the type definitions of `seven5.RestPost` and `seven5.RestPut`.  At run-time, however, the true type of this object will be a pointer to the wire type associated with the resource.  In the case of our sample code for this and the previous chapter, the true type is `*ArticleWire` for the resource implementation `ArticleResource`.  Similarly, the return values all the REST methods specify a type of `interface{}` but the values returned _must_ use the wire type.  _Seven5_ checks the type of the values returned and will abort if the type is not as expected.  The particular check to perform is dictated by the association specified in the call to `BindSeparate()` above.

The author has gone back and forth on whether this use of `interface{}` is counter to the static nature of _Seven5_.  An alternative would be have an interface generation tool that emitted Go source code to correctly and statically type these interfaces.  A hypothetical example with article:

```
type ArticleIndex interface {
        Index(PBundle) (ArticleWire*[], error)
}
```

This option was originally planned, but was eventually discarded for two reasons.  First, there is correctness. It is easy to generate good, fatal errors when the wrong type is returned by a resource to _Seven5_, and resources can do a type conversion to the "correct" wire type with one line of code.  Again, this conversion will abort if the code were not correct.  Second, there are cases where a single resource type may implement two different "logical" REST resources.  Thus, such an implementation in the proposed code generator case, ends up needing the generating a "suffix" to every method, such as `IndexArticleV1` and `IndexArticleV2` to prevent a collision and allow both to be implemented in the same type.  This seems unnecessarily cumbersome and wordy.  The jury remains in deliberation on the type-safety issue here. 

### Practice: Starting the server

The line in `main()`:

```go
http.ListenAndServe(fmt.Sprintf(":%d", heroku.Port()), mux)
```

The [ListenAndServe](http://golang.org/pkg/net/http/#ListenAndServe) call is defined by the [net/http](http://golang.org/pkg/net/http/) package of the standard Go library.  It starts the server listening on the port supplied.  It does not return, instead waiting forever on HTTP requests and dispatching them according to the bindings previously configured into mux.  In the case of our example here, the server would respond to the default _Seven5_ URL bindings (`DefaultProjectBindings()`) and a few URLs that begin with `/rest/article` that have been configured into the dispatcher `bd`.

### Theory: NAME

The name of the application "nullblog" is bound to the constant `NAME` at the top of the `main()` snippet above.  It is used three times early in the `main()` function, and each time use was carefully considered to see if it could be avoided.  In each case, it was sufficiently important to the _called_ code to know the name of the application being run that this parameter was preserved.  It is valid criticism that this is a violation of the Don't Repeat Yourself (DRY) principle, but mitigating this somewhat is the fact that there are valid, more complex, cases where the calling `main()` code might want to supply different values instead of the same value three times.

