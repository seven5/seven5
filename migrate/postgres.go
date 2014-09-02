package migrate

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"io"
	"time"
)

type postgresMigrator struct {
	db *sql.DB
}

//DestroyMigrationRecords destroys the records kept about migrations. This
//is almost certanily a bad idea anywhere except in a test.
func (m *postgresMigrator) DestroyMigrationRecords() error {
	_, err := m.db.Exec(m.deleteAll())
	return err
}

//Close severs the db connection, if it exists.
func (m *postgresMigrator) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

//NewPostgresMigrator returns a new migrator capable of performing a sequence
//of migrations.  Note that this typically will fail (returning the
//error when the DB connection cannot be established).  There is no
//connection pooling or other optimizations done in Migrator because it
//is designed to be used primarily in one-off migration situations.
func NewPostgresMigrator(url string) (Migrator, error) {
	//fmt.Printf("url to parse: %s\n", url)
	opts, err := pq.ParseURL(url)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("opts that resulted: %s\n", opts)

	db, err := sql.Open("postgres", opts)
	if err != nil {
		return nil, err
	}
	result := &postgresMigrator{
		db: db,
	}
	if err := result.createTableIfNotExists(); err != nil {
		return nil, err
	}
	return result, nil
}

//createTableIfNotExists creates the table if it doesn't yet exist
//otherwise does nothing.
func (m *postgresMigrator) createTableIfNotExists() error {
	_, err := m.db.Exec(m.createTable())
	return err
}

//CurrentMigration returns the current migration number found in the
//migrations table. If no migrations have been performed, this returns
//zero.
func (m *postgresMigrator) CurrentMigrationNumber() (int, error) {
	curr := -9018
	row := m.db.QueryRow(m.numQuery())
	err := row.Scan(&curr)
	if err != nil && err != sql.ErrNoRows {
		return -1917, err
	}
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return curr, nil
}

//Goes up to the last known migration. Returns the number of migrations
//successfully performed.  Migrations are done in a transaction.  This
//is shorthand for UpTo(len(migrations),migrations).
func (m *postgresMigrator) Up(migrations map[int]MigrationFunc) (int, error) {
	return m.UpTo(len(migrations), migrations)
}

//Goes Up to the migration given.Returns the number of migrations
//successfully performed.  Migrations are done in a transaction. If
//the current migration level is equal to or greater than the target
//this is a noop.  Note that you pass where you want to BE after this
//is done, not the migration to be performed.  For example, passing
//1 as target implies running migration 1.
func (m *postgresMigrator) UpTo(target int, migrations map[int]MigrationFunc) (int, error) {

	curr, err := m.CurrentMigrationNumber()
	if err != nil {
		return 0, err
	}
	success := 0
	for i := 0; i < target; i++ {
		if i < curr {
			continue
		}
		fmt.Printf("[migrator] attempting migration UP %03d\n", i+1)
		err := m.step(migrations[i+1])
		if err != nil {
			return success, err
		}
		success++
		_, err = m.db.Exec(m.insertRow(i + 1))
		if err != nil {
			return success, err
		}
	}
	return success, nil
}

//Goes down to zero. Returns the number of number migrations done successfully.
//Migrations are done in a transaction.  This is short for
//DownTo(0,migrations).
func (m *postgresMigrator) Down(migrations map[int]MigrationFunc) (int, error) {
	return m.DownTo(0, migrations)
}

//Goes down to zero. Returns the number of number migrations done successfully.
//Migrations are done in a transaction. Note that you pass the destination
//migration number.  You pass 1 to mean that you want migrations 2 to current
//run in reverse order.
func (m *postgresMigrator) DownTo(target int, migrations map[int]MigrationFunc) (int, error) {
	curr, err := m.CurrentMigrationNumber()
	if err != nil {
		return 0, err
	}
	success := 0
	for i := curr; i > target; i-- {
		fmt.Printf("[migrator] attempting migration DOWN %03d\n", i)
		err := m.step(migrations[i])
		if err != nil {
			return success, err
		}
		success++
		_, err = m.db.Exec(m.deleteRow(i))
		if err != nil {
			return success, err
		}
	}
	return success, nil
}

//Single step, given a function.  There is no need for f
func (m *postgresMigrator) step(fn MigrationFunc) error {
	if fn == nil {
		panic("step called but no function!")
	}
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		fmt.Printf("got an error trying to run a step: %v", err)
		if err := tx.Rollback(); err != nil {
			panic(fmt.Sprintf("unable to rollback migration transaction:%v", err))
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		panic(fmt.Sprintf("unable to commit migration transaction:%v", err))
	}
	return nil
}

func (m *postgresMigrator) DumpHistory(writer io.Writer) error {
	rows, err := m.db.Query(m.historyQuery())
	if err != nil {
		return err
	}
	for rows.Next() {
		var n int
		var t time.Time

		err := rows.Scan(&n, &t)
		if err != nil {
			return err
		}
		fmt.Fprintf(writer, "%03d %v\n", n, t)
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	return nil
}

//make it easier to port to other DB
func (m *postgresMigrator) createTable() string {
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (n INTEGER, t TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp)",
		MIGRATION_TABLE)
}

//make it easier to port to other DB
func (m *postgresMigrator) numQuery() string {
	return fmt.Sprintf("SELECT n FROM %s ORDER BY t DESC LIMIT 1", MIGRATION_TABLE)
}

//make it easier to port to other DB
func (m *postgresMigrator) historyQuery() string {
	return fmt.Sprintf("SELECT n,t FROM %s ORDER BY t DESC", MIGRATION_TABLE)
}

//make it easier to port to other DB
func (m *postgresMigrator) deleteAll() string {
	return fmt.Sprintf("DELETE FROM %s", MIGRATION_TABLE)
}

//make it easier to port to other DB
func (m *postgresMigrator) insertRow(i int) string {
	return fmt.Sprintf("INSERT INTO %s (n) VALUES(%d)", MIGRATION_TABLE, i)
}

func (m *postgresMigrator) deleteRow(i int) string {
	return fmt.Sprintf("DELETE FROM %s WHERE n = %d", MIGRATION_TABLE, i)
}
