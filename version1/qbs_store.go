package seven5

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/coocood/qbs"
)

type QbsStore struct {
	Policy *QbsDefaultOrmTransactionPolicy
	Dsn    *qbs.DataSourceName
}

// NewQbsStoreFromDSN creates a *QbsStore from a DSN; DSNs can be created
// directly with ParamsToDSN or from the environment with GetDSNOrDie, via
// the DATABASE_URL environment var.
func NewQbsStoreFromDSN(dsn *qbs.DataSourceName) *QbsStore {
	qbs.RegisterWithDataSourceName(dsn)
	result := &QbsStore{
		Dsn:    dsn,
		Policy: NewQbsDefaultOrmTransactionPolicy(),
	}
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
		if u := os.Getenv("PGUSER"); u != "" {
			user = u
		} else {
			user = "postgres"
		}
	}
	dsn := &qbs.DataSourceName{}
	dsn.DbName = dbname
	dsn.Dialect = StringToDialect(driver)
	dsn.Host = "localhost"
	dsn.Port = "5432"
	dsn.Username = user
	return dsn
}

//This function returns the datasource name for the DATABASE_URL
//in the environment. If the value cannot be found, this panics.
func GetDSNOrDie() *qbs.DataSourceName {
	db := os.Getenv("DATABASE_URL")
	if db == "" {
		panic("no DATABASE_URL found, cannot connect to a *QbsStore")
	}
	u, err := url.Parse(db)
	if err != nil {
		panic(fmt.Sprintf("unable to parse database URL: %s", err))
	}
	log.Printf("DATABASE_URL found, connecting to %+v\n", u)
	dsn := &qbs.DataSourceName{}
	dsn.DbName = u.Path[1:]
	dsn.Host = u.Host
	if strings.Index(u.Host, ":") != -1 {
		pair := strings.Split(u.Host, ":")
		if len(pair) != 2 {
			panic("badly formed host:port")
		}
		dsn.Host = pair[0]
		dsn.Port = pair[1]
	}
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
	panic(fmt.Sprintf("unable to deal with db dialect provided %s", n))
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
		if err.Error() == "EOF" {
			log.Printf("It's likely there is something listening on your server port that isn't the database you expected.")
		}
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
