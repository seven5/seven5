### Example Project For Seven5, "myproj"

#### This Directory, "myproj/dart/myproj/web"

The details of how to layout a Dart project are in pub's [Package Layout](http://pub.dartlang.org/doc/package-layout.html) document.

This directory is for *static* files such as CSS and images.  By convention when using Seven5, you should
use a directory called "assets" to hold these files separated by type; we have done so in this
sample project.

This directory is by default mapped to "/" in your application when you use the Seven5 defaults on
the _server_ side so a CSS file usually can be found in "/assets/css/foo.css" via a web client and
from this source directory under "assets/css/foo.css".

Generally, the files served from this directory by a Seven5 server application are "fixed" and thus the 
client-side code can use a 304 return code to speed up page loading.  The exception is the "app"
directory...

#### The App Directory and Dart Web Components

>>>>Warning: This may change! Still under active development!

A sub directory of this directory is the "app" subdirectory.  This subdirectory is the _output_ location
of the tool "dwc" or "Dart Web components Compiler".  When you are creating links in your application,
you should _never_ reference "/app" but rather the targets "logical" name.  So, to create a link to
a web page "/foo" you should use an anchor tag with "/foo" as the target but in reality this web
page will be *on-demand* compiled to use dart web components in "/app/foo" and the user will be redirected
to this page.  If the user's browser does not support Dart natively, again an on-demand compilation
will occur to convert any Dart code needed to javascript code and serve those to the browser.  

When creating a link from "inside" the "app" directory, the same rules still apply.  If a segment of HTML
is being created in a web component that needs a link to "/bar" it should be listed as such and 
allow the on-demand compilation to occur as the redirection is done to "/app/bar".  Do *not* make the
link point directly to "/app/bar" because this prevents the on-demand compilation from occurring.
This machinery, if a bit clunky, allows a development cycle with "just hit reload".

Note that this generally means that Seven5 client-side apps should use absolute paths from 
"/" for generating links, not relative ones because the app in general does not know (nor want to know)
if there is a need to compile web components or javascript.



