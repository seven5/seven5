package migrate

import (
	"database/sql"
)

const MIGRATION_TABLE = "migrations"

//Migrator is a type meant to represent the small table that is kept about the
//migrations performend and the ability to run up/down migrations.  Note that
//this exposes the lower level sql.Tx interface to callers because it is likely
//that migrations are not portable between databases (there are things you want
//to be able to do that are DB specific).  This is in contrast to the use of
//Qbs elsewhere as the _use_ of the database should be db-implementation
//independent.
type Migrator interface {
	DestroyMigrationRecords() error
	Close() error
	CurrentMigrationNumber() (int, error)
	Up(migrations map[int]MigrationFunc) (int, error)
	Down(map[int]MigrationFunc) (int, error)
	UpTo(int, map[int]MigrationFunc) (int, error)
	DownTo(int, map[int]MigrationFunc) (int, error)
}

//This is the function type that the Migrator operates on.  The implementors
//of this type do not need to start or stop transactions, just use the
//transaction provided and error if something goes wrong.
type MigrationFunc func(*sql.Tx) error

//Definitions is a convenient way to hold the up and down migrations in a
//struct.
type Definitions struct {
	Up   map[int]MigrationFunc
	Down map[int]MigrationFunc
}
