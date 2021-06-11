> There is a lot going on here.  Read this document carefully.

# installation and dev setup

## setup tools
Go into the `tools` directory and create symbolic links to the 
various tools, and implicitly versions, that are named in the
`README.md` in that directory.

It *will not work* if you have things installed in your `PATH`
like `node` and `npm`.  They must be linked into the directory `tools`
because that's the way I roll. You can thank me for it later.

## enable script
There is an _enable script_ called `enable-seven5` in this 
directory.  This is where you put shell commands, environment
variables and the like that are specific to your development
efforts. 

> If you set environment variables in your copy of the enable
> script, be sure that you also start the script by _unsetting_ 
> any variables that are related to, or more importantly the opposite
> of, the ones you are setting. 
> This is so the enable script can be run any number of times and
> always results in a *good state*.  Idempotency for the win.

Use an editor to check that `enable-seven5` makes sense for
your system.  Make changes if you must.  

Source the script into your shell, *do not execute it*.  This 
script is designed to be your security blanket: you can do 
`source enable-seven5` at any time from this directory and
the shell you entered that command into is set for development
on this project.  Sometimes this is referred to as just
"enabling seven5" or "enabling the project".

### `tools` directory rationale
You'll notice that the enable script is quite parsimonious with
the directories in the `PATH`.  Notably, it does not include 
`/usr/local/bin`.  This is intentional.  Many people have the
succumbed to the dreaded disease, _package-manager-itis_.  This
is a chronic inflammation of `/usr/local/bin`.  While this 
disease has usually minor symptoms for a *user* of software, it
can be fatal for *developers*.  If you don't know what version of
the tools you are using, _especially in the javascript world where
the tools change versions rapidly and complex dependencies
must be maintained_, you are doomed.  

If there are some tools in `/usr/local/bin` that you cannot live
without in your shell, link them into the tools directory as
explained in `tools/README.md`.  Some good examples of this need are
`goland`, `emacs`,  or other clearly inferior editors.  

If you are super-paranoid--and I am--this "link things into the
tools directory" strategy allows you to
install a particular version of node like say `14.13.1` in *any
random place* in your filesystem and then link the `tools` directory
commands to it. This guarantees, for example, that you get the version of node
you expect and want.  This is _far_ simpler than the use of `nvm`
or similar. 

## the big warning
> Under no circumstances should you take some random bozo on the
> web's advice to use `npm install -g <whatever>`.   The `npm install -g` 
> command installs the package's binary into
> `/usr/local/bin` or similar; this is a key cause for _package-manager-itis_!  
> You don't want that, for the reasons above.  Should you disregard this warning, 
> many kittens will be left in the sun in the desert to die. 
> 
> Instead, take out the `-g` and, if necessary, link the needed
> executable into the tools directory, usually they are in
> `node_modules/.bin/`.

## install node packages
There are approximately 1.8 zillion dependent packages.  You'll
need to download them all one time to get started. They will be
placed into the `node_modules` directory.  If you get concerned,
or you just "have that nagging feeling" you can nuke the `node_modules`
directory and run the command below anytime.

Make sure your npm is linked correctly with `which npm`.This should result
in a file in your tools directory--if it is not use your enable script!  
You install all the npm packages with `npm install` in this directory.  

### installing the rest of the tools
You can do `ls node_modules/.bin` to see what you have wrought.
With these packages all installed, go to `tools/FOLLOWUP.md` and
follow the instructions for installing the remaining command-line
tools you need.

## development overview

When you see a command that is of the form `npm run <foo>` this
means that entry "foo" is in the `package.json` of this project.  There
is a section of the json in `package.json` called `scripts`.  Here is a
snippet:
```json
{
  "scripts": {
    "dev:serve": "NODE_ENV=development webpack serve --config webpack.react.config.js --mode development --env development",
    "test": "jest",
    "format": "prettier --write src/**/*.{ts,tsx}",
    "lintfix": "eslint --fix src/**/*.{ts,tsx}",
    "redux-devtools": "redux-devtools --hostname=localhost --port=8098",
    "react-devtools": "react-devtools"
  }
  
}
```

## running a development version

```shell
> npm run dev:serve
```
> When you start the command above on a Mac running OSX
> you'll probably get a dialog box that pops up and says "Do you want
> to allow this application to accept incoming connections?" or something
> similar.  You do want that, because webservers don't work without 
> that ability.

You will need to leave this program running.  This is a hot-reloading
server that _in principle_ should reload your code into the browser when
you change it.  (I don't have a lot of confidence in this mechanism, but
your mileage may vary.  I always restart the server anytime something wonky 
happens.) This special development webserver that is connected
to webpack and webpack's module analysis code.  This webserver will watch
your copy of the source code, recompile the files that change and if they
compile ok, inject the newly compiled code (really "module") into 
the browser.

## modifying the code

### editing
If you want to modify the code, it is likely that you'll want to
configure your editor to know about typescript and javascript.

#### goland
Goland is known to work correctly and "do all things" to assist
when coding.  To set it up:
* Configure the node settings to point to the same node instance 
  that is linked to your tools directory.  This should also tells
  goland where your version of npm is (since its in the same `bin`
  directory).
  
## XXX
* Do we need to explain the prettier config?
* Eslint config?
* Typescript config?

### (auto) checking your source code
You'll notice that when you use the hot reload functionality of `dev:serve`,
or just `webpack` to build all the things, 
the compilation step includes "linter" check and a "prettiness check."  
Under the covers this is because both kinds of compilation are driven by webpack.
The configuration `webpack.config.js` is used to drive the whole process.

Our webpack configuration forces any compilation of the code (hot or cold) to also do
a check for problems with `eslint` and a check for code formatting with `prettier`.
Both of these programs, like `webpack`, are linked into your `tools` directory if you
want to run them independently of webpack.  The key thing to note is that
using webpack may (will) cause your files to change to be "prettier".
This happens you run any webpack command, but is more obvious when you run
`npm run dev:serve` since it happens every time you save a file in your
editor.

I've never gotten burned by this setup, but you can
change the "fix" option in the `webpack.config.json` if you are
a nervous type.  Without the "fix" option, prettier just logs to the
terminal what changes it wants.  If your editor detects changes to the
underlying filesystem and notifies you when its buffers get
out of date, you should be fine since the changes to the disk files are always only 
cosmetic.

Our `eslint` rules can be found in the file `.eslintrc.js`.  This file is
short because we are largely borrowing the 
[airbnb typescript rules](https://www.npmjs.com/package/eslint-config-airbnb-typescript)
for both typescript and react. The file `.eslintignore` tells eslint what 
directories to ignore for linting, such as `node_modules`.

Our `prettier` configuration is in `.prettierrc.js`.  Again our configuration is
quite small because we are using the default prettier configuration choices.  If
you want, you can use the command `npm run lintfix` to run prettier in "fix mode"
where it will try fix any of the code formatting problems it finds.  This will
modify your source files in place.

## building a production version
* Need to have a way to configure the NODE_ENV and webpack mode variables
* Need to explain how electron does it's packaging

## modifying the build setup

### webpack
The configuration of the various tools that are in use here is fiendishly 
complex.  Broadly, the "driver" or "coordinator" of all parts of the build
is `webpack`.  This command, if you want, can be invoked from your `tools`
directory and that can be useful for testing out different command line 
options or configuration files.

The two `webpack.*.js` configuration files are intended for different webpack
"targets."  A target to webpack is the intended environment for the resulting 
javascript code.  `webpack.electron.js` is targeted at "electron-main" and
`webpack.react.js` is targeted at "electron-renderer".  electron has a strict
separation of the "main" process that controls the interactions with 
the browser (electron) and operating system and the "rendering" process that
draws the actual app inside a window.  

#### babel
You will notice that many of the tools that we are using reference the transpiling
tool [babel](https://babeljs.io/) in their documentation.  Our configuration does 
_not_ use a `.babelrc` or `babel.config.js` as you might expect.  We use 
webpack's built-in ability to configure babel and just allow webpack to invoke 
babel internally. So we are using babel, but the configuration is done in the 
`webpack.*.js` files.

### typescript configuration
We are using typescript 4.1.3 and its primary configuration file
is `tsconfig.json`.  You can run the typescript the compiler standalone
with `tsc` in your `tools` directory. 

When we are building the source code (either for hot reloading or for file
output), webpack runs the typescript compiler.

It is tricky if you look at the `webpack.*.js` you'll see that we are binding
the `.ts` and `.tsx` to a webpack plugin called `babel-loader`.  In practice 
the relationship between babel version 7 and typescript is very complex.  

[Babel and typescript (from 2019)](https://iamturns.com/typescript-babel/)
 
* Are we sure that babel is invoking the `tsc` or handling the complete 
compilation itself? It is definitely reading the `tsconfig.json`. 

#### typescript extension
We are using a proposed typescript extension for class properties. This means
properties, functions, etc.  The form of this extension is something like the
uses of static variables below:
```typescript
class foo {
  static bar:number =0;
  static baz(): number { return foo.bar++;}
}
```

### modules (and headaches)
I really don't understand exactly what module format we are using.  Our
`tsconfig.json` says that the typescript compiler is outputting commonjs
format modules, also called "cjs" or "node" format.  However, webpack *and* 
babel understand all four of the major module formats and can transform one 
into another.  

* What should we do about this?  
  Can somebody figure out all the various combinations?

### bootstrap
The app is built using bootstrap for the basic look and feel. However,
we use the react implementation of bootstrap, not the standard version.

[react-bootstrap](https://react-bootstrap.github.io/)

## electron security
Currently we have a number of seriously broken configuration choices in our
electron app that we need to fix.

* [electron security tutorial](https://electronjs.org/docs/tutorial/security)
* In the console, we see a number of "security problem" messages

## finding things in the source
The "main" process of the electron app is in `src/main/main.ts`.

We use the redux feature slice approach for our code:
[redux structure](https://redux.js.org/tutorials/essentials/part-2-app-structure)

## running tests
You can run tests either with `npm run test` or `jest`, the latter of which is
linked into your `tools` directory.

Tests are located in the `tests` directory and have a structure that is parallel
to the structure of `src`.

Currently, we have a only unit tests but some of these are actually tests of the
particular rendering behavior.  This means we can test both the higher level of
abstraction (components) as well as how the component is presented.  Naturally,
it is definitely a tension about how much to test the details of the HTML output
because if overdone, this leads to fragility.

To get logical coverage of a feature our tests, we test these ways:
* Test that the redux store works properly when presented with different 
  states and different actions. The redux store is the source of truth and all 
  changes to it are done through actions.
* Unit tests that the transformation from current state in redux is correctly
  mapped to properties for display in our components.
* Unit tests that our custom components transform a set of properties provided to them
  into HTML.
* Unit tests that simulate clicks or other user inputs and test that the correct
  actions are generated to the redux store.
  
``` 
User input > redux actions --> redux store --> redux state > custom component > HTML. 
```
