# Tools directory

You'll need to make a symlink to these programs from this directory.
* node
* npm
* npx

Start by linking your `node`, `npx` and `npm` binaries into this directory
based on precisely which version you want to use. For reference,
I use version `14.13.1`, `6.14.8`, and `6.14.8` respectively. 

All three of these tools are typically in the same directory inside your node
installation (one of the many node installs you have, no doubt).  Here are 
the example symlink commands for my local machine.

```shell
> ln -s ~/src/fsdev-mac/tools/node/bin/node .
> ln -s ~/src/fsdev-mac/tools/node/bin/npx .
> ln -s ~/src/fsdev-mac/tools/node/bin/npm . 
```

I've included `npx` in the set of tools to link because it can be helpful
to use its ability to run commands inside npm packages that are *not* installed
locally in `node_modules`.

You should now return to the directory above and continue
with the `README.md` instructions.