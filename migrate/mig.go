package migrate

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"os"
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
	DumpHistory(io.Writer) error
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

//Main should be called from user-level code's main() function to process
//the arguments and run migrations provided. It reads the arguments provided
//on the command line.
func Main(defn *Definitions, m Migrator) {
	var up, down, history bool
	var step int
	flag.BoolVar(&up, "up", false, "indicates up migrations are desired, with no step specified all possible up migrations are run")
	flag.BoolVar(&down, "down", false, "indicates down migrations are desired, with no step specified all possible down migrations are run")
	flag.BoolVar(&history, "history", false, "dumps the migration history to the terminal")
	flag.IntVar(&step, "step", 0, "number of steps to proceed, dont set this if you want all migrations performed")
	flag.Parse()

	current, err := m.CurrentMigrationNumber()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to get migration number: %v", err)
		return
	}

	if !up && !down {
		fmt.Printf("current migration number is %03d\n", current)
		if history {
			if err := m.DumpHistory(os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "unable to get migration history: %v", err)
			}
		}
		return
	}

	if step < 0 {
		fmt.Fprintf(os.Stderr, "negative steps don't make sense, use --down and a positive step value\n")
		return
	}

	var n int
	if up {
		if step == 0 && current == len(defn.Up) {
			fmt.Printf("at last migration, nothing to do\n")
			return
		}
		if step > 0 && step+current > len(defn.Up) {
			fmt.Fprintf(os.Stderr, "maximum migration number is %d (not %d), no migrations performed\n", len(defn.Up), current+step)
			return
		}
		if step == 0 {
			n, err = m.Up(defn.Up)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to do up migrations: %v\n", err)
				return
			}
		} else {
			n, err = m.UpTo(current+step, defn.Up)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to do up migrations: %v\n", err)
				return
			}
		}
		fmt.Printf("%03d UP migrations performed\n", n)
		return
	}
	//must be down
	if step == 0 && current == 0 {
		fmt.Printf("at earliest migration, nothing to do\n")
		return
	}
	if step > 0 && current-step < 0 {
		fmt.Fprintf(os.Stderr, "earliest migration number is 0 (not %d), no migrations performed\n", current-step)
		return
	}
	if step == 0 {
		n, err = m.Down(defn.Down)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to do down migrations: %v\n", err)
			return
		}
	} else {
		n, err = m.DownTo(current-step, defn.Down)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to do down migrations: %v\n", err)
			return
		}
	}
	fmt.Printf("%03d DOWN migrations performed\n", n)
}
