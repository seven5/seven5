--- 
layout: book
chapter: Running Server Tests
---

### Goal: Running server tests
After reading this chapter, you should be able to run server tests and understand how to write tests for the server-side of a _Seven5_ application. For this chapter, you can continue to use the code you checked out in a previous chapter, based on the branch "code-book-1". 

### Practice: Building and running the tests

Seven5 tests are just standard Go tests so they can be run like this:

```
$ go test nullblog
ok  	nullblog	0.022s
```

When run, your tests may take less or more than 0.022 seconds, but should display "ok" to indicate all tests passed.  The command "go test" is smart enough to compile the test code, plus any changed library code before running the tests.

### Practice: How To Write A Simple Test

Below is a simple, silly test of our code in the method `RestFind` on the type `nullblog.ArticleResource`

```go
func TestFind(T *testing.T) {

        underTest := new(ArticleResource)

        result, err := underTest.Find(0, nil)
        if err != nil {
                T.Fatalf("unexpected error in Find: %s", err)
        }
        article := result.(*ArticleWire)
        if article.Id != 0 || article.Author == "" || article.Content == "" {
                T.Fatalf("unexpected content in Find %+v", article)
        }

        result, err = underTest.Find(29892, nil)
        if result != nil {
                T.Fatalf("expected error but got result in Find %+v", result)
        }
        if err == nil {
                T.Fatalf("nothing returned from Find (two nils)!")
        }

        http_error := err.(*seven5.Error)
        if http_error.StatusCode != http.StatusBadRequest {
                T.Errorf("wrong error code, expect BadRequest but got %d", http_error.StatusCode)
        }

}
```

The code tests the object `underTest`, an instance of `nullblog.ArticleResource`.  Since that type holds no state, it is easy to create one to put it under test. The first logical test checks that article 0 can be retrieved without any errors and that it has at least some content inside it.  The second logical test checks that calling `Find()` on the object under test with an out of range Id value (29892) _does_ produce an error.  It further insures that the correct HTTP error code was returned, `http.StatusBadRequest`.   `TestFind()` is testing the exact same code path that would be run in a production setting.  It may be viewed as either a unit test or a functional test, depending on your point of view.


### Theory: If you make it easy to test, they will come

These tests in `/tmp/book/go/src/nullblog/nullblog_test.go`  were intentionally kept simple to illustrate the general idea of testing the implementation of a resource separate from the network particulars.  These tests can be run without the complexity of requiring a copy of the server to be running, a common problem with testing web-based back-ends.  These tests also can be tested synchronously, independent of the particular, probably performance-critical, behavior that they would be subjected to in a production server.

This short chapter should now also help clarify why the _library_ "nullblog" is separated from the _command_ "runnullblog" that creates a working server but whose implementation is entirely inside the library.  The library/command split makes the library testing simpler.









