# Seven5: Develop

<nav>
    <ul>
        <li>[Intro](index.html)</li>
        <li>[Install](install.html)</li>
        <li>[Develop](develop.html)</li>
        <li>[Pontificate](pontificate.html)</li>
    </ul>
</nav>

The following steps will create a basic blog application named "blargh":

## Create the project directory:

    seven5-create-project blargh

This will create the following directory structure:

    blargh
        |
        +-- handlers.go 
        |
        +-- resources.go 
        |
        +-- sass
        |    |
        |    +-- site.sass 
        |
        +-- js
        |    |
        |    +-- site.js 
        |
        +-- mongrel2.sqlite 
        |
        +-- mongrel2
             |
             +-- log
             |    |
             |    +-- access.log 
             |    |
             |    +-- error.log 
             |    |
             |    +-- seven5.log 
             |     
             +-- run
             |    |
             |    +-- mongrel2.pid 
             |    |
             |    +-- run_control (a unix domain socket)
             |
             +-- static
                  |
                  +-- favicon.png


* `handlers.go`, `resources.go` and contain (respectively) the views and data. 

* `mongrel2.sqlite` contains configuration data that is needed for mongrel2 to connect to the application in test mode.

* `log` contains logs (duh) including access, error, and logging from the app code.

* `run` contains control files generated and used by the front end.

* `static` contains files which will be served by mongrel2 unchanged. The web server will be chrooted to this directory.

## Run the server

	cd blargh
	seven5d

## Make a change, feel the love

## Create and serve a blab resource

## Enjoy the free JS, HTML, SASS, Ajax and events

## Style that bad boy

## Squirt it to the cloud
