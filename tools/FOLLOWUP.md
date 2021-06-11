## linking in the other commands

Now that you have installed all the node packages by following
the instructions in `../README.md`, you can put
the rest of the commands into your tools directory (this one).

You should include a command line version of

* eslint
* prettier
* tsc
* webpack

Many of these _can_ be run by other means such as using npx or 
adding scripts to `../package.json`.  If you prefer that, go for it,
I just feel its easier to have all the tools (and their implicit
versions) in one place.

Here are the commands you can use for the linking of the commands that
are downloaded to the `node_modules` directory.

```shell
ln -s ../node_modules/.bin/eslint .
ln -s ../node_modules/.bin/prettier .
ln -s ../node_modules/.bin/tsc .
ln -s ../node_modules/.bin/webpack .
```

See the _Tools Rationale_ section of `../README.md` for why this procedure is
both righteous and just.