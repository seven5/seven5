### Example Project For Seven5 Book

#### This directory, nullblog/go/src/nullblog/runnullblog

This directory contains a typically small "main" entry-point for the server side of a project.  The
name is by convention "runnullblog" for a project called "nullblog".  

You can install the executable to "nullblog/go/bin/runnullblog" with this command:

```
go install nullblog/runnullblog
```

If you have other executable programs that depend on the library "nullblog", or are just organized as a 
part of "nullblog", the convention is to name the _entry-point_ as "nullblog/run<some_support_program>".  
They can thus be installed the same way as "runnullblog" and kept together with the server's executable
in "nullblog/go/bin".

If you need to have support scripts of various kinds, they should also be installed "nullblog/go/bin" or
just in "nullblog" if they are dependent on the current working directory.