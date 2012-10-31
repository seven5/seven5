--- 
layout: page
title: For The Impatient
tagline: "Don't read the directions."
---
# Seven5

## An opinionated, stiff web framework written in Go and Dart

* Seven5 is RESTful with remorse or pity.
* Seven5 is fiercely reactionary towards the forces of dynamism.

## For the impatient, start really fast here

### Install Prerequisites

* [Go](http://golang.org/doc/install)
* [Dartium](http://www.dartlang.org/dartium/)
* [Git](https://help.github.com/articles/set-up-git)

### Get the example app: 'italy'

{% highlight console %}
$ mkdir /tmp/foo
$ cd /tmp/foo
$ git clone -b examples git@github.com:seven5/seven5.git examples
{% endhighlight %}

### Setup environment

{% highlight console %}
$ cd examples/example1/go
$ export GOPATH=`pwd`
$ export PATH=$PATH:$GOPATH/bin

{% endhighlight %}

### Get seven5tool and seven5 library sources, build everything

{% highlight console %}
$ go get github.com/seven5/seven5tool
{% endhighlight %}

### Build sample app

{% highlight console %}
$ go install italy/runitaly
$ which runitaly 
{% endhighlight %}

### Run sample app, server part

{% highlight console %}
$ runitaly 
{% endhighlight %}

### Run Dartium 

The example below is the version of the command to run Dartium on OS X.  It's better, but not mandatory, if you run it with `enable_type_checks` and `enable_asserts` turned on, as this enforces the type checking.

{% highlight console %}
$ DART_FLAGS='--enable_type_checks --enable_asserts' /path/to/dart/../Chromium.app/Contents/MacOS/Chromium 
{% endhighlight %}


### Check out the app

You have to use Dartium to go to this web.

{% highlight console %}
$ open http://localhost:3003/static/italy
{% endhighlight %}

#### Screenshot of running app

The screenshot below shows the output of `http://localhost:3003/static/italy.html`.  Under the covers, this is running GET on the collection resource, `/italiancity/` in the URL space, then formatting the JSON result into HTML.

![Italy Application Screengrab](/seven5/images/italy-snap.png)

### Play with it to see more RESTful love

* Try deleting a city  (DELETE)
* Try adding a city.  The values in the from fields are sanity checked (so lat and long must be in-bounds).  (POST)
* Try clicking on a city name to change it.  (PUT)

### Check out the actual message traffic

* With the focus on the application window, view the javascript console.  In this case, it's really the Dart console: `View > Developer > Javascript Console`.

* Click on the network icon 

* Hit Reload on Dartium browser

![Javascript Console Screengrab](/seven5/images/console-snap.png)

##### JSON Payloads

You can click on any of the objects in the far left column to get details about the particular messages sent to and received from the sever.  You can examine the json payloads as well.  For example, here is a snap of of the GET payload on `/italiancity/`:

![JSON display Screengrab](/seven5/images/json-snap.png)

### Turn on google maps

Assuming you have a google API key and have "Static Maps API" ( *not* the "Maps API") turned on, you can try this:

* Make sure your google API key is enabled for `localhost:3003/*`
 
* Edit `go/src/italy/publicsetting.json` and follow the instructions in the file.

* Restart the server by running `runitaly` again.  `publicsettings.json`, like most things in the _Seven5_ world is read once at startup and the results cached.

Here's what the output looks like when you have the google maps API turned on:

![Italy with maps Screengrab](/seven5/images/withmaps-snap.png)

### The "wire" structure of an italian city

In the file `go/src/italy/city.go` the structure that is passed over the wire is defined.  This is the type, called the _wire type_ in _Seven5_, that is input and output in the URL space at `italiancity`:

{% highlight go %}
//sub structure used for a latitude, longitude pair
type LatLng struct {
  Latitude  seven5.Floating
  Longitude seven5.Floating
}

//rest wire type for a single city, properties must be public for JSON encoder
type ItalianCity struct {
  Id         seven5.Id
  Name       seven5.String255
  Population seven5.Integer
  Province   seven5.String255
  Location   *LatLng
}

{% endhighlight %}

You'll note that all the types must come from the seven5 definitions, such as `seven5.String255`.  This is to ease the translation between Go, Json, SQL, and Dart.  The set of types that can be used is *severely* restricted to prevent any confusion over the type mappings.  Note that this restriction applies _only_ to wire types, not implementation types.  Implementation types are usually called _resource types_ in _Seven5_ and typically are named `FooResource` if they are the implementation of the wire type `Foo`. 

#### The generated dart code

You can browse to `http://localhost:3003/generated/dart`  to see the client-side code that was generated from the structs above.

This enables code like this on the client side (be sure to note the types):
{% highlight java %}
//
// Load all cities that are known from the server (index translates to GET /italiancity/)
//
...
	ItalianCity.Index(dumpAll);
...

//callback from the query that gets all cities (1st parameter is a list)
void dumpAll(List<ItalianCity> cities, HttpRequest result) {
	print("number of cities returned from Index: ${cities.length}");
	for (ItalianCity city in cities) {
		print("    city returned from Index(): [${city.Id}] ${city.Name}");
	}
	print("result of 'Index' (GET): ${result.status} ${result.statusText}");

}

{% endhighlight %}

or like this:

{% highlight java %}
//
// Load a specific instance from the server (find translations to GET /italiancity/2)
//
ItalianCity genoa = new ItalianCity();
genoa.Find(2, (city, result) {
	assert(city.Name=="Genoa");
});
genoa.Find(16, null, (result) {
	assert(result.status==400);
});
{% endhighlight %}

### Use 'seven5tool' to create a new project

This assumes that you followed the directions above in terms of setting environment variables and such.

{% highlight console %}
$ mkdir /tmp/bar
$ cd /tmp/bar
$ seven5tool bigidea myproject

project 'myproject' created.  You probably want to set GOPATH and PATH like this:
export GOPATH=/tmp/bar/myproject/go:$GOPATH
export PATH=$PATH:/tmp/bar/myproject/go/bin

{% endhighlight %}

Follow the tools recommendation and reset your `GOPATH` and `PATH` for development of myproject.

### Build the sample project created

{% highlight console %}
$ go install myproject/runmyproject
$ cd myproject/go/src
$ ls
{% endhighlight %}

#### Begin hacking. Enjoy.