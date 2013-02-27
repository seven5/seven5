### Example Project For Seven5, "myproj"

#### This directory, myproj/go/src/myproj/runmyproj

This directory contains a typically small "main" entry-point for the server side of a project.  The
name is by convention "runmyproj" for a project called "myproj".  

You can install the executable to "myproj/go/bin/runmyproj" with this command:

```
go install myproj/runmyproj
```

If you have other executable programs that depend on the library "myproj", or are just organized as a 
part of "myproj", the convention is to name the _entry-point_ as "myproj/run<some_support_program>".  
They can thus be installed the same way as "runmyproj" and kept together with the server's executable
in "myproj/go/bin".

If you need to have support scripts of various kinds, they should also be installed "myproj/go/bin" or
just in "myproj" if they are dependent on the current working directory.