package main

import (
	"flag"
	"github.com/seven5/seven5"
	"nullblog"
)

func main() {
	env := seven5.NewLocalhostEnvironment("nullblog", false /*not test*/)
	migrator := seven5.NewQbsMigrator(env.GetQbsStore(), true, false)
	target := migrator.ParseMigrationFlags(flag.NewFlagSet("myflags", flag.PanicOnError))
	migrator.Migrate(target, &nullblog.Migrate{})
}
