package seven5

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/coocood/qbs"
	//"log"
	"os"
	"reflect"
	"strings"
	"time"
)

//Migrator is an interface describing an API for a utility type,
//primarily used when writing migrations for apps or when testing.
type Migrator interface {
	//FindMigrationsInOrder is primarily of interest to Migrator implementors.
	//It should not be need for normal application development.
	FindMigrationsInOrder(int, int, interface{}) []*MigrationInfo
	//DetermineCurrentMigrationNumber is primarily of interest to Migrator implementors.
	//It should not be need for normal application development.
	DetermineCurrentMigrationNumber() (int, error)
	//CreateMigrationTable is primarily of interest to Migrator implementors.
	//It should not be need for normal application development.
	CreateMigrationTable() error
	//RunMigrations is primarily of interest to Migrator implementors
	//It should not be need for normal application development.
	RunMigrations(fn []*MigrationInfo) error

	//CreateTableIfNotExists should be used by applications that wish to create
	//a new table.
	CreateTableIfNotExists(interface{}) error
	//DropTableIfExists should be used by applications that wish to destroy
	//an existing table.
	DropTableIfExists(interface{}) error
	//ParseMigrations adds the flags to have a "standard" interface to running
	//migrations from the command line.
	ParseMigrationFlags(fset *flag.FlagSet) int
	//Migrate is the entry point for driving the migrations defined for this
	//application.  Typically it's called by migration or test code.  The
	//parameter specifies the target state or UNLMIMITED_MIGRATIONS.  The latter
	//parameter will be interrogated for migration methods.
	Migrate(int, interface{}) error

	//ToZero is shorthand Migrate(0,fns)
	ToZero(interface{}) error
	//ToMax is shorthand Migrate(UNLIMITED_MIGRATIONS,fns)
	ToMax(interface{}) error
}

//BaseMigrator is the part of the migrator code that is useful to any implementation
//of Migrator.
type BaseMigrator struct {
	Verbose bool
}

//QbsMigrator is an implementation of Migrator that implements the Migrator
//interface and specifically
type QbsMigrator struct {
	*BaseMigrator
	Store *QbsStore
}

const (
	//UNLIMITED_MIGRATIONS means that all migrations that can be found in
	//forward direction are desired.
	UNLIMITED_MIGRATIONS = 10000
)

//MigrationInfo is used to describe a migration and have enough information for
//giving the human an understanding of what has been done.
type MigrationInfo struct {
	v           reflect.Value
	n           int
	is_rollback bool
}

//MigrationRecord is a recording of the processing of a migration that is stored
//in the database by this infrastructure.
type MigrationRecord struct {
	Id      int64
	N       int
	Created time.Time
}

//FindMigrationsInOrder returns a list of migrations that are needed to reach
//the provided target.  Migration functions are discovered on m by looking for
//the pattern "Migrate_0001" and "Migrate_0001_Rollback" in function names.
func (self *BaseMigrator) FindMigrationsInOrder(target int, i int, migrationContainer interface{}) []*MigrationInfo {

	limit := 100000
	increment := 1
	suffix := ""

	result := []*MigrationInfo{}

	if i > target {
		increment = -1
		suffix = "_Rollback"
		limit = target
	} else if i < target {
		i++
	} else {
		if self.Verbose {
			fmt.Printf("Nothing to do, at correct migration (%04d).\n", i)
		}
		return nil
	}

	v := reflect.ValueOf(migrationContainer)
	for i != limit {

		//found no more migrations in the sequence
		expected := fmt.Sprintf("Migrate_%s%s", fmt.Sprintf("%04d", i), suffix)
		if self.Verbose {
			fmt.Printf("Looking for migration %s...", expected)
		}
		fn := v.MethodByName(expected)
		if !fn.IsValid() {
			if self.Verbose {
				fmt.Printf("not found (%v).\n", fn.Kind())
			}
			break
		}
		if self.Verbose {
			fmt.Printf("found it.\n")
		}
		//found something, add to result
		result = append(result, &MigrationInfo{fn, i, increment < 0})

		i += increment

		//reached stopping condition?
		if i == target {
			break
		}
	}
	if len(result) > 0 {
		direction := "forward"
		if increment < 0 {
			direction = "backward"
		}
		if self.Verbose {
			fmt.Printf("%04d %s migrations needed.\n", len(result), direction)
		}
	}
	return result

}

//CreateMigrationTable creates, if necessary, the table that holds the migration
//records for this database.  Note that this is database specific.
func (self *QbsMigrator) CreateMigrationTable() error {
	migration, err := qbs.GetMigration()
	if err != nil {
		return err
	}
	defer migration.Close()
	err = migration.CreateTableIfNotExists(&MigrationRecord{})
	return err
}

//DetermineCurrentMigrationNumber reads the database for the most recent MigrationRecord
//and returns the migration number of that record, or 0 if no MigrationRecord
//table can be found.
func (self *QbsMigrator) DetermineCurrentMigrationNumber() (int, error) {
	err := self.CreateMigrationTable()
	if err != nil {
		return -881, err
	}
	qbs, err := qbs.GetQbs()
	if err != nil {
		return -881, err
	}

	var r MigrationRecord
	qbs.OrderByDesc("created")
	if err := qbs.Find(&r); err != nil {
		if err == sql.ErrNoRows {
			if self.Verbose {
				fmt.Printf("No previous migration records found.\n")
			}
			return 0, nil
		}
		return -881, err
	}
	if self.Verbose {
		fmt.Printf("Last update was migration %04d at %v\n", r.N, r.Created)
	}
	return r.N, nil
}

func (self *QbsMigrator) RunMigrations(info []*MigrationInfo) error {

	for _, pair := range info {
		result := pair.v.Call([]reflect.Value{reflect.ValueOf(self)})
		if len(result) != 1 {
			errmsg := fmt.Sprintf("unable to understand result (%d items returned)", len(result))
			panic(errmsg)
		}
		if !result[0].IsNil() {
			err, ok := result[0].Interface().(error)
			if !ok {
				errmsg := fmt.Sprintf("unable to understand result of calling migration_%04d (%s)",
					pair.n, result[0])
				panic(errmsg)
			}
			return err
		}

		q, err := qbs.GetQbs()
		if err != nil {
			return err
		}
		defer q.Close()

		current := pair.n
		if pair.is_rollback {
			current--
		}
		rec := &MigrationRecord{N: current}
		_, err = q.Save(rec)

		if err != nil {
			return err
		}
		direction := ""
		if pair.is_rollback {
			direction = " (rollback)"
		}
		if self.Verbose {
			fmt.Printf("Applied migration %04d%s.\n", pair.n, direction)
		}
	}
	return nil
}

//ParseMigrationFlag is a convenience to help those writing migrations.  It
//adds some default flags to the flag set to allow the user to specify a
//migration and then parses the program's arguments and
func (self *BaseMigrator) ParseMigrationFlags(fset *flag.FlagSet) int {

	migration := -1

	fset.IntVar(&migration, "m", UNLIMITED_MIGRATIONS, "migration number to change to")

	if err := fset.Parse(os.Args[1:]); err != nil {
		errmsg := fmt.Sprintf("failed to parse arguments: %s", err)
		panic(errmsg)
	}

	if migration < 0 {
		panic("you must supply a migration number with a value of at least 0 with -m flag")
	}

	return migration
}

func (self *QbsMigrator) Migrate(target int, migrationFns interface{}) (result error) {
	defer func() {
		if raw := recover(); raw != nil {
			err := fmt.Sprintf("---> %s", raw)
			if strings.HasSuffix(err, "bad connection") {
				fmt.Fprintf(os.Stderr, "unable to connect to database: check db name, credentials, connectivity, and if the database is running\n")
				result = errors.New("bad connection")
			}
		}
	}()
	curr, err := self.DetermineCurrentMigrationNumber()
	if err != nil {
		return err
	}
	info := self.FindMigrationsInOrder(target, curr, migrationFns)
	if err := self.RunMigrations(info); err != nil {
		return err
	}
	return nil
}

func (self *QbsMigrator) ToZero(migrationFns interface{}) error {
	return self.Migrate(0, migrationFns)
}

func (self *QbsMigrator) ToMax(migrationFns interface{}) error {
	return self.Migrate(UNLIMITED_MIGRATIONS, migrationFns)
}

func (self *QbsMigrator) CreateTableIfNotExists(struct_ptr interface{}) error {
	m, err := qbs.GetMigration()
	if err != nil {
		return err
	}
	defer m.Close()
	return m.CreateTableIfNotExists(struct_ptr)
}

func (self *QbsMigrator) DropTableIfExists(struct_ptr interface{}) error {
	m, err := qbs.GetMigration()
	if err != nil {
		return err
	}
	defer m.Close()
	m.DropTable(struct_ptr)
	return nil
}

//NewQbsMigrator returns a new migrator implementation based on Qbs for the ORM.
func NewQbsMigrator(s *QbsStore, verbose bool, log bool) *QbsMigrator {
	result := &QbsMigrator{BaseMigrator: &BaseMigrator{}, Store: s}
	result.Store.Q.Log = log
	result.Verbose = verbose
	return result
}
