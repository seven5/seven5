
## Commands

* A command has a single, very-specific task.
* The _Seven5_ application proper is built as a collection of commands, each command implemented in Go.
* A command receives a structure as a parameter and returns a structure as a result.
* If you are "taking the defaults", you should never need to worry about commands.
* If you want to change the commands, you can do so per-application with `commands.json` at the root of your application directory.
* `commands.json` is watched by the roadie and he'll rebuild _Seven5_ based on your new commands.
* `ValidateProjectFiles` is an example of a groupie; so is `ExplodeType` from our discussion of pills.
