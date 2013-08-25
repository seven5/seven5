package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/coocood/qbs"
	_ "github.com/lib/pq"
	"nullblog"
	"os"
	"reflect"
	"strings"
	"time"
)

const (
	UNLIMITED_MIGRATIONS = 10000
)

func getVar(appname string, varname string) string {
	return os.Getenv(strings.ToUpper(appname + "_" + varname))
}

func parseMigrationFlags(fset *flag.FlagSet, appname string) (*qbs.DataSourceName, int) {

	var dbname, dbuser, dbpass, driver, dbhost, dbport string
	migration := -1

	fset.StringVar(&driver, "driver", getVar(appname, "driver"),
		"sets the database type driver (experimental)")

	fset.StringVar(&dbname, "dbname", getVar(appname, "dbname"),
		"sets the database name connected to; "+
			"use only to explicitly override the default environment variable")
	fset.StringVar(&dbuser, "dbuser", getVar(appname, "dbuser"),
		"sets the database user name to connected with; "+
			"use only to explicitly override the default environment variable")
	fset.StringVar(&dbpass, "dbpass", getVar(appname, "dbpass"),
		"sets the database password to connected with; "+
			"use only to explicitly override the default environment variable")
	fset.StringVar(&dbhost, "dbhost", getVar(appname, "dbhost"),
		"sets the database host to connected to; "+
			"use only to explicitly override the default environment variable")
	fset.StringVar(&dbport, "dbport", getVar(appname, "dbport"),
		"sets the database port to connected to; "+
			"use only to explicitly override the default environment variable")
	fset.IntVar(&migration, "m", UNLIMITED_MIGRATIONS, "migration number to change to")

	if err := fset.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse arguments: %s", err)
	}

	if migration < 0 {
		panic("you must supply a migration number with a value of at least 0 with -m flag")
	}

	if dbname == "" {
		errmsg := fmt.Sprintf("you must supply a database name or have the environment variable %s_%s set",
			strings.ToUpper(appname), "DBNAME")
		panic(errmsg)
	}

	var dsn *qbs.DataSourceName

	if driver == "postgres" || driver == "" {
		dsn = qbs.DefaultPostgresDataSourceName(dbname)

		//apply the db variables they have set
		if dbuser != "" {
			dsn.Username = dbuser
		}
		if dbpass != "" {
			dsn.Password = dbpass
		}
		if dbhost != "" {
			dsn.Host = dbhost
		}
		if dbport != "" {
			dsn.Port = dbport
		}

	} else if driver == "sqlite3" {
		dsn = new(qbs.DataSourceName)
		dsn.DbName = dbname
		dsn.Dialect = qbs.NewSqlite3()
	} else {
		errmsg := fmt.Sprintf("unable to understand driver %s", driver)
		panic(errmsg)
	}

	return dsn, migration
}

type migrationAndNumberPair struct {
	v           reflect.Value
	n           int
	is_rollback bool
}

func findMigrationsInOrder(target int, m interface{}) []*migrationAndNumberPair {

	i, err := determineCurrentMigrationNumber()
	if err != nil {
		errmsg := fmt.Sprintf("unable to determine migration number, "+
			"did you forget to create the database? (%s)", err)
		panic(errmsg)
	}

	if target < 0 {
		target = UNLIMITED_MIGRATIONS
	}

	limit := 100000
	increment := 1
	suffix := ""

	result := []*migrationAndNumberPair{}

	if i > target {
		increment = -1
		suffix = "_Rollback"
		limit = target
	} else if i < target {
		i++
	} else {
		fmt.Printf("nothing to do, at correct migration (%04d)\n", i)
	}

	v := reflect.ValueOf(m)
	for i != limit {

		//found no more migrations in the sequence
		fn := v.MethodByName(fmt.Sprintf("Migrate_%s%s", fmt.Sprintf("%04d", i), suffix))
		if !fn.IsValid() {
			break
		}

		//found something, add to result
		result = append(result, &migrationAndNumberPair{fn, i, increment < 0})

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
		fmt.Printf("%04d %s migrations needed...\n", len(result), direction)
	}
	return result

}

func runMigrations(dsn *qbs.DataSourceName, fn []*migrationAndNumberPair) error {

	for _, pair := range fn {
		migration, err := qbs.GetMigration()
		if err != nil {
			return err
		}
		defer migration.Close()

		result := pair.v.Call([]reflect.Value{reflect.ValueOf(migration)})
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

		rec := &MigrationRecord{N: pair.n}
		_, err = q.Save(rec)
		if err != nil {
			return err
		}
		direction := ""
		if pair.is_rollback {
			direction = "(rollback)"
		}
		fmt.Printf("applied migration %04d %s\n", pair.n, direction)
	}
	return nil
}

type MigrationRecord struct {
	Id      int64
	N       int
	Created time.Time
}

func createMigrationTable() error {
	migration, err := qbs.GetMigration()
	if err != nil {
		return err
	}
	defer migration.Close()
	err = migration.CreateTableIfNotExists(&MigrationRecord{})
	return err
}

func determineCurrentMigrationNumber() (int, error) {
	err := createMigrationTable()
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
			fmt.Printf("No previous migration records found.\n")
			return 0, nil
		}
		return -881, err
	}
	fmt.Printf("Last update was migration %04d at %v\n", r.N, r.Created)
	return r.N, nil
}

type NullBlogMigrate struct {
	//no need to hold any state
}

func (self *NullBlogMigrate) Migrate_0001(m *qbs.Migration) error {
	return m.CreateTableIfNotExists(&nullblog.Article{})
}

func (self *NullBlogMigrate) Migrate_0001_Rollback(m *qbs.Migration) error {
	m.DropTableIfExists(&nullblog.Article{})
	return nil
}

func main() {
	fset := flag.NewFlagSet("nullblog_migrate_flags", flag.PanicOnError)
	dsn, target := parseMigrationFlags(fset, "nullblog")
	qbs.RegisterWithDataSourceName(dsn)
	migrations := findMigrationsInOrder(target, &NullBlogMigrate{})
	runMigrations(dsn, migrations)
}
