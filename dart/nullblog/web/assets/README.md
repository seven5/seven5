### Example Project For Seven5, "myproj"

#### This Directory, "nullblog/dart/nullblog/web/assets"

The details of how to layout a Dart project are typically only related to the _parent_ directories 
of this one.  These details are in pub's 
[Package Layout](http://pub.dartlang.org/doc/package-layout.html) document.

This directory is fixed, slowing changing asset files such as CSS or images.  By convention, these are
placed in the directories "css" and "img".  In Go, the Seven5 code can access these files via the
type `ProjectFinder`.  This directory becomes "/assets" from the point of view of a web client.

In the unlikely event that you need HTML code to be returned by the server for fixed web content, this
should be served from the "html" child directory of this directory.  Be aware that for fixed web content,
Seven5 provides alternate "go level" mechanisms such as "embed file" with `seven5tool` that allows 
fixed content to reside directly in the application's binary.

>>>> In the future, `seven5tool` will include a mechanism to load _all_ fixed content into the binary of the application so that applications can be run entirely without a filesystem.