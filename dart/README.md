### Example Project For Seven5, "myproj"

#### This Directory, "myproj/dart"

This is the root directory of the client side of a project called "myproj".  In a strict sense, Dart
does not expect "myproj" to be a directory of that name; we have done that here with the child of
this directory.  This choice was made to allow multiple, client-side "apps" to share the same
installation (this directory's children) and back-end (in the "myproj/go" heirarchry).  

This is a blatant rip-off of the strategy used successfully by Django with their 
"multiple apps in one project" approach.