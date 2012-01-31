package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"seven5"
	"errors"
)

var pathErr = errors.New("the parent of the project directory must be called 'src' and that must be in a directory inside GOPATH")

//at the moment, you can't use go to build something that's not correctly housed in
//some part of go path
func checkGOPATH(directory string) error {
	
	gopath:=filepath.SplitList(os.Getenv("GOPATH"))
	
	dir:=filepath.Dir(directory)
	
	if filepath.Base(dir) !="src"{
		return pathErr
	}
	
	candidate := filepath.Dir(dir)
	
	for _, p:=range gopath {
		if filepath.Clean(candidate) == filepath.Clean(p) {
			return nil
		}
	}
	
	return pathErr
}

//validateEnvVars checks the environment variables and returns a map of error messages--mapping
//failed environemnt variable name to happy message. Empty map returned if everyhing is ok.
func validateEnvVars() map[string]string {

	err := make(map[string]string)

	GOOS := os.Getenv("GOOS")
	GOARCH := os.Getenv("GOARCH")
	GOROOT := os.Getenv("GOARCH")
	GOBIN := os.Getenv("GOBIN")

	if GOOS == "" {
		err["GOOS"] = "the name of the target operating system you are developing for, such as 'linux', 'darwin', 'freebsd' or 'windows'.  See http://golang.org/doc/install.html"
	}
	if GOARCH == "" {
		err["GOARCH"] = "the name of the target processor architecture you are developing for, such as '386', 'amd64', or 'arm'.  See http://golang.org/doc/install.html"
	}
	if GOBIN == "" {
		err["GOBIN"] = "the directory that contains your standard go binaries are located, typically $GOROOT/bin.  See http://golang.org/doc/install.html"
	}
	if GOROOT == "" {
		err["GOROOT"] = "the root directory of your go installation.  See http://golang.org/doc/install.html"
	}

	if len(err) > 0 {
		return err
	}

	GOPATH := os.Getenv("GOPATH")
	if GOPATH == "" {
		err["GOPATH"] = "the root directory of your third-party go package collection. See http://code.google.com/p/go-wiki/wiki/GOPATH"
	}

	return err
}

//validateBinaries checks that the binaries we expect for a proper configuration are present in
//the path. it returns an error map if something is wrong, or an empty map if everything is
//ok.  the error map has the name of the problem binary as the key and the error message as
//the value.
func validateBinaries() map[string]string {
	err := make(map[string]string)

	if _, e := exec.LookPath("go"); e != nil {
		err["go"] = "package build and install tool.  See http://weekly.golang.org/cmd/go/"
	}
	if _, e := exec.LookPath("go"); e != nil {
		err["git"] = "source code management tool, used to download some static parts of a new project, nothing else.  http://git-scm.com/download"
	}
	if _, e := exec.LookPath("mongrel2"); e != nil {
		err["mongrel2"] = "webserver that seven5 uses to serve static content and manage your application. Mongrel2 must be in your path.  see http://mongrel2.org"
	}
	if _, e := exec.LookPath("memcached"); e != nil {
		err["memcached"] = "storage service used by this version of seven5. should be running at all times and if it's not in your path, you probably don't have it installed.  see http://memcached.org/"
	}

	if len(err) > 0 {
		return err
	}

	if _, e := exec.LookPath("rock_on"); e != nil {
		err["rock_on"] = "Seven5's continuous build tool.  See http://seven5.github.com/seven5/install.html"
	}
	if _, e := exec.LookPath("tune"); e != nil {
		err["tune"] = "Seven5's code generator for building 'main' functions for user programs.  See http://seven5.github.com/seven5/install.html"
	}
	if _, e := exec.LookPath("big_idea"); e != nil {
		err["big_idea"] = "Seven5's tool for setting up a new project.  You are running it now, but it should probably be in your path."
	}

	return err
}

//buildDirectoryStructure builds the necessary Seven5 directory structure and creates all the 
//static files thatare not dependent on the users chosen type.
func buildProjectStructure(projectName string, dirPath string) error {
	log := filepath.Join(dirPath, "mongrel2", "log")
	run := filepath.Join(dirPath, "mongrel2", "run")
	static := filepath.Join(dirPath, "mongrel2", "static")
	css := filepath.Join(dirPath, "mongrel2", "static", "css")
	images := filepath.Join(dirPath, "mongrel2", "static", "images")
	js := filepath.Join(dirPath, "mongrel2", "static", "js")

	fm:=os.FileMode(0775)
	
	if err := os.MkdirAll(log,fm); err != nil {
		return err
	}

	if err := os.MkdirAll(run,fm); err != nil {
		return err
	}

	if err := os.MkdirAll(static,fm); err != nil {
		return err
	}

	if err := os.MkdirAll(css,fm); err != nil {
		return err
	}

	if err := os.MkdirAll(images,fm); err != nil {
		return err
	}

	if err := os.MkdirAll(js,fm); err != nil {
		return err
	}

	content := fmt.Sprintf(`This is the primary directory for developing your back-end go code.
This directory will have a %s_test.sqlite file once you begin running your code 
inside mongrel2. This directory is watched by the rock_on program during normal development,
usually started with a command like 'rock_on %s'
`, projectName, projectName)

	if err := createFile(filepath.Join(dirPath, "README"), content); err != nil {
		return err
	}

	content = fmt.Sprintf(`package %s
		
import (
	"log"
	"seven5"
	"seven5/store"
)
//You should put code in this file if you need to do private initialization of some kind.
//"Private" initialization is anything that you do not want to check into the source code 
//repository, such as initialization of variables with passwords, creating super-users, etc.
//
//The function must have exactly the signature below.  The program 'tune' will scan this file
//for this function and, if found, it will hook it to the Seven5 infrastructure so it gets
//called at application start-up time.
//
//Be sure to check that your .svnignore, .gitignore or other configuration files are correctly
//set to have your version control system not assume you want to commit this file.
//
//PrivateInit is run to allow an application to initalize datastructures or the store without
//needing to check the file into the repository.  This is called at application start-up
//time. If PrivateInit returns an error, program execution will cease.
func PrivateInit(log *log.Logger, store store.T) error {
	
	//used by the tests generated by big_idea tool
	_,err:= seven5.CreateUser(store,"joe","joe","smith","joe@example.com","joe"/*password!!*/)
	return err
}
`, projectName)

	if err := createFile(filepath.Join(dirPath, "pwd.go"), content); err != nil {
		return err
	}

	content = "This directory contains the mongrel2 and seven5 log files.\n"
	if err := createFile(filepath.Join(dirPath, "mongrel2", "log", "README"), content); err != nil {
		return err
	}

	if err := createFile(filepath.Join(dirPath, "mongrel2", "log", "access.log"), ""); err != nil {
		return err
	}

	if err := createFile(filepath.Join(dirPath, "mongrel2", "log", "mongrel2.err.log"), ""); err != nil {
		return err
	}

	if err := createFile(filepath.Join(dirPath, "mongrel2", "log", "mongrel2.out.log"), ""); err != nil {
		return err
	}

	if err := createFile(filepath.Join(dirPath, "mongrel2", "log", "seven5.log"), ""); err != nil {
		return err
	}

	content = `This directory contains control files for mongrel2 such as the pid file (mongrel2.pid)
and the unix domain socket that Seven5 uses to communicate with mongrel2 (control)
`

	if err := createFile(filepath.Join(dirPath, "mongrel2", "run", "README"), content); err != nil {
		return err
	}

	content = `This directory is the one that mongrel2 is chrooted to.  However, only the
directory 'static' has been entered into the set of routes mongrel2 will consider, so
the directories run and log are not visible from a browser.
`

	if err := createFile(filepath.Join(dirPath, "mongrel2", "README"), content); err != nil {
		return err
	}

	content = fmt.Sprintf(`This directory contains all the javascript and other 'static' resources that are served
up by mongrel2 on behalf of your application.  This directory is visible in URL space from a 
browser as /static. 

There are two special subdirectories here, 'seven5js' and 'vendorjs'.  These two directories are
maintained as separate repositories within the seven5 project
(https://github.com/seven5/vendorjs and https://github.com/seven5/seven5js) 
and as git submodules.  It is important to realize that when a project like %s is created, the
big_idea tool creates a copy (snapshot) of these repositories.  It is possible to upgrade these
two directories if seven5 changes. 
`, projectName)

	if err := createFile(filepath.Join(dirPath, "mongrel2", "static", "README"), content); err != nil {
		return err
	}

	content = `Your CSS files go here.  We recommend using 'lesscss' which is part of the
vendorjs package supplied to this project.  See http://lesscss.org for more info. 
`

	if err := createFile(filepath.Join(dirPath, "mongrel2", "static", "css", "README"), content); err != nil {
		return err
	}

	content = fmt.Sprintf(`<!--- your static HTML content -->

	<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN"
	  "http://www.w3.org/TR/html4/loose.dtd">
	<html>
	<head>
	  <title>%s</title>

	  <!--  
	  #### Javascript Libraries Needed For Working With Seven5  #####
	   -->

	  <!--jquery -->
	  <script type="text/javascript" src="vendorjsjs/jquery-1.7/jquery-1.7.js"></script>

	  <!--backbone + underscore for client side MVC-->
	  <script type="text/javascript" src="vendorjsjs/underscore-1.2.4/underscore.js"></script>
	  <script type="text/javascript" src="vendorjsjs/backbone-0.5.3/backbone.js"></script>

	  <!-- Seven5 support library -->
	  <script type="text/javascript" src="seven5js/seven5-0.1.js"></script>

	  <!--lesscss stylesheets-->
	  
	  <!--
	  <link rel="stylesheet/less" href="css/%s.less" type="text/css" media="screen" charset="utf-8">
	  -->
	
	  <!--lesscss code to transform the less to css-->
	  <script type="text/javascript" src="vendorjs/lesscss-1.1.6/less-1.1.6.js"></script>

	  <!--  
	  #### APPLICATION CODE  #####
	   -->

	  <!--seven5 models-->
	  <script type="text/javascript" src="/api/seven5/models"></script>

	  <!--your application-->
	  <script type="text/javascript" src="js/%s.js"></script>
	</head>
	<body>
	</body>
	</html>

`, projectName, projectName, projectName)

	if err := createFile(filepath.Join(dirPath, "mongrel2", "static", "index.html"), content); err != nil {
		return err
	}

	content = `Your images go here. 
`

	if err := createFile(filepath.Join(dirPath, "mongrel2", "static", "images", "README"), content); err != nil {
		return err
	}

	content = `This is the primary directory for your application's client side code in
javascript.   You should also put application tests in this directory.
`

	if err := createFile(filepath.Join(dirPath, "mongrel2", "static", "js", "README"), content); err != nil {
		return err
	}

	content = fmt.Sprintf(`
	<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN"
	  "http://www.w3.org/TR/html4/loose.dtd">
	<html>
	<head>
	  <title>%s Test Runner</title>

	  <!--  
	  #### Javascript Libraries Needed For Working With Seven5  #####
	   -->

	  <!--to style the test results-->
	  <link rel="stylesheet" type="text/css" href="vendorjs/jasmine-1.1.0/jasmine.css">

	  <!--jasmine for testing and displaying results -->
	  <script type="text/javascript" src="vendorjs/jasmine-1.1.0/jasmine.js"></script>
	  <script type="text/javascript" src="vendorjs/jasmine-1.1.0/jasmine-html.js"></script>

	  <!--jquery -->
	  <script type="text/javascript" src="vendorjs/jquery-1.7/jquery-1.7.js"></script>

	  <!--for jquery testabiliciousness -->
	  <script type="text/javascript" src="vendorjs/jasmine-jquery/lib/jasmine-jquery.js"></script>

	  <!--backbone + underscore for client side MVC-->
	  <script type="text/javascript" src="vendorjs/underscore-1.2.4/underscore.js"></script>
	  <script type="text/javascript" src="vendorjs/backbone-0.5.3/backbone.js"></script>

	  <!--we do these loads explicitly to avoid problems with the requirejs cruft that sinon uses-->
	  <script src="vendorjs/Sinon.JS-1.3.1/lib/sinon.js"></script>
	  <script src="vendorjs/Sinon.JS-1.3.1/lib/sinon/util/event.js"></script>
	  <script src="vendorjs/Sinon.JS-1.3.1/lib/sinon/util/fake_xml_http_request.js"></script>
	  <script src="vendorjs/Sinon.JS-1.3.1/lib/sinon/util/fake_server.js"></script>
	  <script src="vendorjs/Sinon.JS-1.3.1/lib/sinon/util/fake_timers.js"></script>

	  <script src="vendorjs/Sinon.JS-1.3.1/lib/sinon/spy.js"></script>
	  <script src="vendorjs/Sinon.JS-1.3.1/lib/sinon/stub.js"></script>
	  <script src="vendorjs/Sinon.JS-1.3.1/lib/sinon/mock.js"></script>
	  <script src="vendorjs/Sinon.JS-1.3.1/lib/sinon/collection.js"></script>
	  <script src="vendorjs/Sinon.JS-1.3.1/lib/sinon/sandbox.js"></script>

	  <!-- Seven5 support library -->
	  <script type="text/javascript" src="seven5js/seven5-0.1.js"></script>

	  <!--  
	  #### APPLICATION CODE  #####
	   -->

	  <!--lesscss stylesheets-->
	  <!--
	  <link rel="stylesheet/less" href="css/%s.less" type="text/css" media="screen" charset="utf-8">
	  -->
	  <!--lesscss code to transform the less to css-->
	  <script type="text/javascript" src="vendorjs/lesscss-1.1.6/less-1.1.6.js"></script>

	  <!--for parsing URLs, part of testing -->
	  <script type="text/javascript" src="vendorjs/jquery-url-parser-2.0/jquery-url-2.0.js"></script>

	  <!--seven5 models-->
	  <script type="text/javascript" src="/api/seven5/models"></script>

	  <!--application under test-->
	  <script type="text/javascript" src="js/%s.js"></script>

	  <!--tests-->
	  <script type="text/javascript" src="js/%s-test.js"></script>
	</head>
	<body>

	<script type="text/javascript">
	  jasmine.getEnv().addReporter(new jasmine.TrivialReporter());
	  jasmine.getEnv().execute();
	</script>


	</body>
	</html>
`, projectName, projectName, projectName, projectName)

	if err := createFile(filepath.Join(dirPath, "mongrel2", "static", "SpecRunner.html"), content); err != nil {
		return err
	}

	return nil
}

//up returns the first character as upper case.
func up(s string) string {
	return strings.ToUpper(s[0:1])+s[1:]
}

func buildComplexFiles(projectName, dirPath, typeName string) error {
	
	content:=fmt.Sprintf(`
package %s

import (
	"seven5"
	"time"
) 

type %s struct {
	Name string ` +"`seven5key:\"Name\"`"+`
	Code int
	Id uint64 ` + "`json:\"id\"`"+`  //has to be lower case for Backbone.js
	When time.Time
}

func New%sSvc() seven5.Httpified{
	//2nd param: true means you don't need to sign in to read values
	//3rd param: false means anybody who is signed in can change things
	return seven5.ScaffoldRestService(make([]*%s,0,0),  true, false)
}
`, projectName,up(typeName),up(typeName),up(typeName));

	if err := createFile(filepath.Join(dirPath, typeName+".bbone.go"), content); err != nil {
		return err
	}
	
	content=fmt.Sprintf(`
package %s

//This is a place to put server-side tests.  You can test here or by using the client-side
//RestTester.  This is probably a better place to test if you have very complex server-side
//logic.
//
//The tests link.bbone_test.go in the dungheap sample program may give you some ideas.
//
//Note: If you are testing on the server side, you need to have memcached running or
//the tests will always fail!

import (
	"launchpad.net/gocheck"
	//"seven5"
	"seven5/store"
	"testing"
)

// Hook up gocheck into the default gotest runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

// This is the "suite" structure for objects that need to survive the whole of the tests
// in this file.
type %sSuite struct {
	Impl  store.StoreImpl
	Store store.T
}

var suite = &%sSuite{}

// hook up suite to gocheck
var _ = gocheck.Suite(suite)

//we create a conn to the memcached at start of the suite
func (self *%sSuite) SetUpSuite(c *gocheck.C) {
	self.Impl = store.NewStoreImpl(store.MEMCACHE_LOCALHOST)
	self.Store = store.NewGobStore(self.Impl)
}

//no need to destry the connection, the program is ending anyway
func (self *%sSuite) TearDownSuite(c *gocheck.C) {
}

//before each test 
func (self *%sSuite) SetUpTest(c *gocheck.C) {
	err := self.Impl.DestroyAll(store.MEMCACHE_LOCALHOST)
	if err != nil {
		c.Fatal("unable to setup test and clear memcached")
	}
}

//after each test
func (self *%sSuite) TearDownTest(c *gocheck.C) {
}

//test your logic or whatever here
func (self *%sSuite) TestYourTest(c *gocheck.C) {
}
`,projectName,typeName,typeName,typeName,typeName,typeName,typeName,typeName)

	if err := createFile(filepath.Join(dirPath, typeName+".bbone_test.go"), content); err != nil {
		return err
	}

	
	
	content = `
	//
	//your javascript application code goes in this file
	//
	`
	if err := createFile(filepath.Join(dirPath, "mongrel2", "static", "js", projectName+".js"), content); err != nil {
		return err
	}

	content = fmt.Sprintf(restTests,seven5.Pluralize(typeName),projectName, 
	seven5.Pluralize(typeName), up(typeName), seven5.Pluralize(typeName),typeName)

	if err := createFile(filepath.Join(dirPath, "mongrel2", "static", "js", projectName+"-test.js"), content); err != nil {
		return err
	}
	return nil
	
}
//createFile creates the file at the path given and writes the content.  It returns immediately
//if any of the operations has an error.
func createFile(path string, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	return f.Close()
}

//pullGitSubprojects gets the two git subprojects into the right place in the new project
//to be useful.
func pullGitSubprojects(directory string) error{
	
	vendorjs:=exec.Command("git", "clone", "git://github.com/seven5/vendorjs.git", filepath.Join(directory,"mongrel2","static","vendorjs"))
	if err:=vendorjs.Run(); err!=nil {
		return err
	}

	seven5js:=exec.Command("git", "clone", "git://github.com/seven5/seven5js.git", filepath.Join(directory,"mongrel2","static","seven5js"))
	if err:=seven5js.Run(); err!=nil {
		return err
	}
	return nil
}

func main() {

	err := validateEnvVars()
	if len(err) > 0 {
		fmt.Fprintf(os.Stderr, "Your environment variables do not appear to be configured properly.\n")
		for k, v := range err {
			fmt.Fprintf(os.Stderr, "The environment variable '%s' should be set to \n%s\n", k, v)
		}
		return
	}

	err = validateBinaries()
	if len(err) > 0 {
		fmt.Fprintf(os.Stderr, "Your 'PATH' environment variable do not appear to be configured properly.\n")
		for k, v := range err {
			fmt.Fprintf(os.Stderr, "The executable program '%s' could not be found.  It is needed because it is\n%s\n", k, v)
		}
		_, hasBigIdea := err["big_idea"]
		if len(err) > 1 || !hasBigIdea {
			return
		}
	}

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "%s: two required arguments, a directory to be created and a structure type to create.\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s: the structure type should be a lower case and singular noun, in english.\n", os.Args[0])
		return
	}

	directory := os.Args[1]
	noun:=os.Args[2]

	pwd, e := os.Getwd()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%s: unable to read current working directory\n", os.Args[0])
		return
	}
	directoryPath := filepath.Join(pwd, directory)
	directoryPath = filepath.Clean(directoryPath)

	file, e := os.Open(directoryPath)
	if file != nil {
		fmt.Fprintf(os.Stderr, "%s: directory '%s' already exists, not touching anything.\n", os.Args[0], directoryPath)
		return
	}

	if e:=checkGOPATH(directoryPath); e!=nil {
		fmt.Fprintf(os.Stderr, "%s: To build a project you have to have a properly configured GOPATH\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s: and within one of the elements of that path, you should run this\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s: tool inside a directory called 'src'.  Nothing done.\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s: Error was %v.\n", os.Args[0],e)
		return
	}

	if e := buildProjectStructure(directory, directoryPath); e != nil {
		fmt.Fprintf(os.Stderr, "%s: unable to create entire directory structure!\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s: error during creation was %v!\n", os.Args[0], e)
		fmt.Fprintf(os.Stderr, "%s: you may want to delete the leftover directory '%s' now\n", os.Args[0], directoryPath)
		return
	}

	if e := buildComplexFiles(directory, directoryPath, noun); e != nil {
		fmt.Fprintf(os.Stderr, "%s: unable to complete construction of the new project!\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s: error during creation was %v!\n", os.Args[0], e)
		fmt.Fprintf(os.Stderr, "%s: you may want to delete the leftover directory '%s' now\n", os.Args[0], directoryPath)
		return
	}
	
	if e:= pullGitSubprojects(directory); e!=nil {
		fmt.Fprintf(os.Stderr, "%s: unable to use git to create the seven5 and vendor javascript!\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s: error running git was %v!\n", os.Args[0], e)
		fmt.Fprintf(os.Stderr, "%s: you may want to delete the leftover directory '%s' now\n", os.Args[0], directoryPath)
		return
	}
}


var restTests = `
//
// This is a demo of how to test your rest services from javascript.  This assumes that your
// services are running on the same machine as the test and thus the responses will be fast.
//
// This assumes that you have not modified the default %s created in %s/pwd.go

$("body").ajaxError(function(event,request,settings) {
	console.log("AJAX ERROR!");
	console.log(event);
	console.log(request);
	console.log(settings);
	throw "no sense continuing, we had an ajax error"
});

var maxWaitTime = 300;  //milliseconds


//
// Tests for the REST service for %s... be VERY aware that these tests are all run
// concurrently.  It is a VERY BAD IDEA to try to use data from one test in another because
// there will be race conditions all over the shop.  Also, be very careful to clean up
// the data after make changes because otherwise the next run of the tests is likely
// to blow up.
//

describe("%s REST service", function() {
	//the service under test
	var exampleSvc=seven5.restService("/api/%s");
	//the tester object
	var RT=seven5.createRestTester("%s",maxWaitTime);
	
	var e1 = {
		Name:"doofus", 
		Code: 25, 
		Time: 0
	};

	var e2 = {
		Name:"fred", 
		Code: 30, 
		Time: 0
	};

	var e3 = {
		Name:"samuel", 
		Code: 100, 
		Time: 0
	};

	describe("create an example", function() {
		it("should allow create with login--then delete of same", function () {
			RT.loginAndRun("joe","joe",
			function(result,sessionId) {
				exampleSvc.create(e1, RT.success,sessionId);
			},
			function(result,sessionId) {
				expect(result.Name).toEqual(e1.Name)
				expect(result.Code).toEqual(e1.Code)
				exampleSvc.delete(result.id, RT.success, sessionId)
				},
			function(result,sessionId) {
				expect(result.Name).toEqual(e1.Name)
				expect(result.Code).toEqual(e1.Code)
			}); 
		});//it
		it("should not allow create without login", function () {
			RT.loginRunAndExpectError("joe","joe",
			function(result,sessionId) {
				exampleSvc.create(e1, RT.success/* omit sessionId so no login)*/)
			});
		});//it
	}); //describe create
	describe("search", function() {
		it("should return empty results when searching for something not there", function() {
			RT.loginAndRun("joe","joe",
				function(result, session) {
					var p={};;
					p["Name"]="moses";
					exampleSvc.fetch(p,RT.success, session);
				},
				function(result,session) {
					//no moses but no error
					expect(result.length).toEqual(0);
					expect(result.error).not.toBeDefined();
			});
		}); //it
		it("should return an error if you search for something not a key", function() {
			RT.loginRunAndExpectError("joe","joe",
				function(result, session) {
					var p={};;
					p["When"]="whenIsNotAKey";
					exampleSvc.fetch(p,RT.success, session);
					});
		}); //it
		it("should return empty results when searching for something not there, even not logged in", function() {
			RT.loginAndRun("joe","joe",
				function(result, session) {
					var p={};;
					p["Name"]="ezekial";
					exampleSvc.fetch(p,RT.success /* no sessionId, so not logged in*/);
				},
				function(result,session) {
					//no ezekial, but no error
					expect(result.length).toEqual(0);
					expect(result.error).not.toBeDefined();
			});
		}); //it
		it("should return one result if we create it first", function() {
			var foundId
			RT.loginAndRun("joe","joe",
				function(result, sessionId) {
					exampleSvc.create(e2, RT.success,sessionId);
				},
				function(result,sessionId) {
					//check that create went ok, then search by name
					expect(result.id).toBeDefined();
					foundId=result.id
					var p={};;
					p["Name"]=e2.Name
					exampleSvc.fetch(p,RT.success,sessionId);
				},
				function(result,sessionId) {
					expect(result.length).toEqual(1)
					expect(result[0].id).toBeDefined();
					expect(result[0].id).toEqual(foundId)
					expect(result[0].Name).toEqual(e2.Name)
					expect(result[0].Code).toEqual(e2.Code)
					expect(result[0].When).toEqual("0001-01-01T00:00:00Z")
					
					//after we are done, kill this example or next test run will fail
					jasmine.getEnv().currentSpec.after(function() { 
						console.log("deleting example created in search test");
						exampleSvc.delete(foundId,RT.success,sessionId);
					});
					
			});
		}); //it
	}); //search
	describe("update", function() {
		it("should allow update if logged in", function() {
			var samId;
			//tricky bit: if you want to allow people to use keys that have special
			//characters in them--like dot, space--you need to do work to make sure
			//to "mangle" the keys in your service.  since we are using scaffolding
			//we can't do that here.
			var changeName="SamuelLJackson";
			
			RT.loginAndRun("joe","joe",
				function(result, sessionId) {
					exampleSvc.create(e3, RT.success,sessionId);
				},
				function(result,sessionId) {
					//check that create went ok, then search by name
					expect(result.id).toBeDefined();
					samId=result.id
					var p={};
					p["Name"]=changeName
					exampleSvc.update(samId,p,RT.success,sessionId);
				},
				function(result,sessionId) {
					expect(result.id).toEqual(samId)
					expect(result.Name).toEqual(changeName)
					expect(result.Code).toEqual(e3.Code)
					expect(result.When).toEqual("0001-01-01T00:00:00Z")
					
					//after we are done, rever back to not having him in database
					jasmine.getEnv().currentSpec.after(function() { 
						console.log("deleting example created in update test");
						exampleSvc.delete(samId,RT.success,sessionId)
					});
				});
			});//it
	}); //describe update
}); //describe example
`