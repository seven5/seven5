This set of functional tests expects that you have two servers running.

* You need to have a copy of the webserver that has the code under test.
** The easiest way to get this is to clone the github repo for seven5 and _select the gh-pages branch_, like this ` git clone https://github.com/seven5/seven5.git -b gh-pages`.
** Then cd into the newly created directory `cd seven5/todomvc`
** From here you need to run a webserver capable of serving the static files in the directory.  
*** A simple webserver called 'ws' is provided with seven5 for this purpose. 
*** If you your GOPATH and PATH set properly for development with seven5, you can do "go install github.com/seven5/ws" to create that server in your $GOPATH/bin directory.  
*** You can run that webserver by just calling the program `ws`. It always uses the current directory and serves static content from that directory.  It also will recognize the special prefix "GOPATH" to look for source files in your GOPATH directories.
** The server must be running on port 8898, which is the default for the ws server.

* You need to have a webdriver-compatible selenium driver running.
** The webdriver service must be visible on port 4444.
** We recommend using [ChromeDriver](https://code.google.com/p/selenium/wiki/ChromeDriver) because it is easy to see what it is doing.
** With ChromeDriver installed, you should start it like this: `./chromedriver --verbose --port=4444`
** It is not known why the tests fail using phantomjs 1.9.7 (which includes a webdriver-compatible server).  The problem may be timing related or due to changes in the phantomjs webdriver.
*** To try phantomjs, you need to change the name of the browser expected in the source in the function `newRemote` to "phantomjs" from "chrome".
*** You can run phantomjs like this: `phantomjs --webdriver=4444`

* You can run the functional tests against the two servers above like this `go test github.com/seven5/seven5/todomvc/functest`
** You will notice that a few Chrome windows are being created and destroyed by chrome driver as the tests run. One window is created for each call to newRemote() in the source.
