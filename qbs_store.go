package seven5

import (
	"flag"
	"fmt"
	"github.com/coocood/qbs"
	"os"
)

const (
	UNLIMITED_MIGRATIONS = 10000
)

type QbsStore struct {
	Q      *qbs.Qbs
	Policy *QbsDefaultOrmTransactionPolicy
	Dsn    *qbs.DataSourceName
}

//NewQbsStore creates an instance of the object that is used to represent a
//connection to the database via the Qbs ORM.  This function will panic if
//the database cannot be correctly connected; this is usually what you want
//because you can't do much if you can't get to the database.
func NewQbsStore(env *EnvironmentVars) *QbsStore {
	result := &QbsStore{}
	result.Dsn = result.DataSourceFromEnvironment(env)
	result.Policy = NewQbsDefaultOrmTransactionPolicy()
	qbs.RegisterWithDataSourceName(result.Dsn)
	q, err := qbs.GetQbs()
	if err != nil {
		panic(err)
	}
	result.Q = q
	return result
}

//DataSourceFromEnvironment creates a correctly formed DataSourceName for use by
//Qbs from the EnvironmentVars supplied.  This function will panic if there is
//no way to retreive the dbname to connect to.
func (self *QbsStore) DataSourceFromEnvironment(env *EnvironmentVars) *qbs.DataSourceName {
	var dsn *qbs.DataSourceName

	driver := env.GetAppValue("driver")
	dbname := env.MustAppValue("dbname")
	dbuser := env.GetAppValue("dbuser")
	dbpass := env.GetAppValue("dbpass")
	dbhost := env.GetAppValue("dbhost")
	dbport := env.GetAppValue("dbport")

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

	return dsn
}

//ParseMigrationFlag is a convenience to help those writing migrations.  It
//adds some default flags to the flag set to allow the user to specify a
//migration and then parses the program's arguments and
func ParseMigrationFlags(fset *flag.FlagSet, env *EnvironmentVars) (*QbsStore, int) {

	migration := -1

	store := NewQbsStore(env)

	fset.IntVar(&migration, "m", UNLIMITED_MIGRATIONS, "migration number to change to")

	if err := fset.Parse(os.Args[1:]); err != nil {
		errmsg := fmt.Sprintf("failed to parse arguments: %s", err)
		panic(errmsg)
	}

	if migration < 0 {
		panic("you must supply a migration number with a value of at least 0 with -m flag")
	}

	return store, migration
}

//QbsDefaultOrmTransactionPolicy is a simple implementation of transaction
//policy that is sufficient for most applications.
type QbsDefaultOrmTransactionPolicy struct {
}

//NewQbsDefaultOrmTransactionPolicy returns a new default implementation of policy
//that will Rollback transactions if there is a 400 or 500 returned by the client. It will also
//rollback if a non-http error is returned, or if the called code panics.  After rolling
//back the transaction, it allows the panic to continue.
func NewQbsDefaultOrmTransactionPolicy() *QbsDefaultOrmTransactionPolicy {
	return &QbsDefaultOrmTransactionPolicy{}
}

//StartTransaction returns a new qbs object after creating the transaction.
func (self *QbsDefaultOrmTransactionPolicy) StartTransaction(q *qbs.Qbs) *qbs.Qbs {
	if err := q.Begin(); err != nil {
		panic(err)
	}
	return q
}

//HandleResult determines whether or not the transaction provided should be rolled
//back or if it should be committed.  It rolls back when the result value is
//a non-http error, if it is an Error and the status code is >= 400.
func (self *QbsDefaultOrmTransactionPolicy) HandleResult(tx *qbs.Qbs, value interface{}, err error) (interface{}, error) {
	if err != nil {
		switch e := err.(type) {
		case *Error:
			if e.StatusCode >= 400 {
				rerr := tx.Rollback()
				if rerr != nil {
					return nil, rerr
				}
			}
		}
	} else {
		if cerr := tx.Commit(); cerr != nil {
			return nil, cerr
		}
	}
	return value, err
}

//HandlePanic rolls back the transiction provided and then panics again.
func (self *QbsDefaultOrmTransactionPolicy) HandlePanic(tx *qbs.Qbs, err interface{}) {
	if rerr := tx.Rollback(); rerr != nil {
		panic(rerr)
	}
	panic(err)
}
