
#### Why do we use pills?

* Go compiles _really_ fast.
* Go's source code is complex to analyze, particularly the type system.
* Go has excellent facilities for reflection.
* Go doesn't support dynamic loading (nor should it).

Taken together, we can simulate the development cycle of a dynamic, interpreted language using pills. How?

* You change your code in `foo.go`, say, introducing some new type `bar`.
* The Roadie notices the file changed, and tries to compile your library.  If this fails, he shows you the error in your browser.  Repeat to yourself, "everything is strongly typed."  Thank you.
* If your library compiles, the Roadie notifies _Seven5_ that changes have been made to the types.
* _Seven5_ sends a message back to the Roadie about new (or changed) types, such as `bar`.
* The Roadie messages _Seven5_ to (re-)inform him about the new types and _Seven5_ generates a bunch of Go code with names like `bar_networking.go` and `bar_storage.go`
* Roadie compiles the resulting library, now with extra code courtesy of _Seven5_, wrapped around `bar`
* Road hooks `bar` up to your web browser so you can access it.

This whole process takes less than a second on a modern laptop.
