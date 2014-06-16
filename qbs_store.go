package seven5

import (
	"fmt"
	"github.com/coocood/qbs"
	"log"
	"net/http"
	"net/url"
	"os"
)

type QbsStore struct {
	Q      *qbs.Qbs
	Policy *QbsDefaultOrmTransactionPolicy
	Dsn    *qbs.DataSourceName
}

// NewQbsStore creates an instance of the object that is used to represent a
// connection to the database via the Qbs ORM.
// NewQbsStore creates a *QbsStore from a DSN; DSNs can be created
// directly with ParamsToDSN or from the environment EnvironmentUrlToDSN (via
// the DATABASE_URL environment var).
func NewQbsStore(dsn *qbs.DataSourceName) *QbsStore {
	result := &QbsStore{}
	result.Dsn = dsn
	result.Policy = NewQbsDefaultOrmTransactionPolicy()
	qbs.RegisterWithDataSourceName(result.Dsn)
	q, err := qbs.GetQbs()
	if err != nil {
		panic(err)
	}
	result.Q = q
	return result
}

//ParamsToDSN allows you to create a DSN directly from some values. This
//is useful for testing.  If driver or user is "", the default driver and
//user are used.
func ParamsToDSN(dbname string, driver string, user string) *qbs.DataSourceName {
	if driver == "" {
		driver = "postgres"
	}
	if user == "" {
		user = "postgres"
	}
	dsn := &qbs.DataSourceName{}
	dsn.DbName = dbname
	dsn.Dialect = StringToDialect(driver)
	dsn.Username = user
	return dsn
}

//This function returns the datasource name for the DATABASE_URL
//in the environment. If the value cannot be found, this panics.
func EnvironmentUrlToDSN() *qbs.DataSourceName {
	db := os.Getenv("DATABASE_URL")
	if db == "" {
		panic("no DATABASE_URL found, cannot connect to a *QbsStore")
	}
	log.Printf("found a database:%s", db)
	u, err := url.Parse(db)
	if err != nil {
		panic(fmt.Sprintf("unable to parse database URL: %s", err))
	}
	log.Printf("got a url %+v\n", u)
	dsn := &qbs.DataSourceName{}
	dsn.DbName = u.Path
	dsn.Host = u.Host
	p, set := u.User.Password()
	if set {
		dsn.Password = p
	}
	dsn.Username = u.User.Username()
	dsn.Dialect = StringToDialect(u.Scheme)
	return dsn
}

//StringToDialect returns an sql dialect for use with QBS given a string
//name.  IF the name is not known, this code panics.
func StringToDialect(n string) qbs.Dialect {
	switch n {
	case "postgres":
		return qbs.NewPostgres()
	case "sqlite3":
		return qbs.NewSqlite3()
	}
	panic(fmt.Sprintf("unable to deal with db dialact provided %s", n))
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
		default:
			log.Printf("got an error type that wasnt HTTP specific, rolling back and returning 500 to client (%v)\n", err)
			rerr := tx.Rollback()
			if rerr != nil {
				return nil, rerr
			}
			return nil, HTTPError(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		}
	} else {
		if cerr := tx.Commit(); cerr != nil {
			return nil, cerr
		}
	}
	return value, err
}

//HandlePanic rolls back the transiction provided and then panics again.
func (self *QbsDefaultOrmTransactionPolicy) HandlePanic(tx *qbs.Qbs, err interface{}) (interface{}, error) {
	log.Printf("got panic, rolling back and returning 500 to client (%v)\n", err)
	if rerr := tx.Rollback(); rerr != nil {
		panic(rerr)
	}
	return nil, HTTPError(http.StatusInternalServerError, fmt.Sprintf("panic: %v", err))
}

//QbsDefaultOrmTransactionPolicy is a simple implementation of transaction
//policy that is sufficient for most applications.
type QbsDefaultOrmTransactionPolicy struct {
}

func WithEmptyQbsStore(store *QbsStore, migrations interface{}, fn func()) {
	migrator := NewQbsMigrator(store, false, false)
	defer func() {
		migrator.Store.Q.Close()
	}()
	migrator.ToZero(migrations)
	migrator.ToMax(migrations)
	fn()
}
