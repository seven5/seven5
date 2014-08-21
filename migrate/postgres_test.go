package migrate

import (
	"database/sql"
	"fmt"
	"testing"
)

const DB_NAME = "seven5test"

//you have two bogeys on your 6...
var migs = Definitions{
	Up: map[int]MigrationFunc{
		1: oneUp,
		2: twoUp,
	},
	Down: map[int]MigrationFunc{
		1: oneDown,
		2: twoDown,
	},
}

//
// MIGRATION ONE: UP
//
func oneUp(tx *sql.Tx) error {
	_, err := tx.Exec("CREATE TABLE foobar (i int, s varchar(255))")
	return err
}

//
// MIGRATION TWO: UP
//
func twoUp(tx *sql.Tx) error {
	_, err := tx.Exec("ALTER TABLE foobar ADD COLUMN t text")
	if err != nil {
		return err
	}
	rows, err := tx.Query("SELECT i,s FROM foobar")
	if err != nil {
		return err
	}
	forLater := make(map[int]string)
	for rows.Next() {
		var i int
		var s string
		if err := rows.Scan(&i, &s); err != nil {
			return err
		}
		forLater[i] = s
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i, s := range forLater {
		//Does this work with ? placeholders instead of sprintf?
		_, err := tx.Exec(fmt.Sprintf("UPDATE foobar SET t='%s' WHERE i = %d", s, i))
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec("ALTER TABLE foobar DROP COLUMN s")
	if err != nil {
		return err
	}

	return nil
}

//
// MIGRATION TWO: DOWN
//
func twoDown(tx *sql.Tx) error {
	_, err := tx.Exec("ALTER TABLE foobar ADD COLUMN s varchar(255)")
	if err != nil {
		return err
	}
	rows, err := tx.Query("SELECT i,t FROM foobar")
	if err != nil {
		return err
	}
	forLater := make(map[int]string)
	for rows.Next() {
		var i int
		var t string
		if err := rows.Scan(&i, &t); err != nil {
			return err
		}
		forLater[i] = t
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i, t := range forLater {
		///XXX Why does this not work with ? as placeholders for the values?
		_, err := tx.Exec(fmt.Sprintf("UPDATE foobar SET s = '%s' WHERE i = %d", t, i))
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec("ALTER TABLE foobar DROP COLUMN t")
	if err != nil {
		return err
	}
	return nil
}

//
// MIGRATION ONE: DOWN
//
func oneDown(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE foobar")
	return err
}

func TestOpen(t *testing.T) {
	_, err := NewPostgresMigrator(DB_NAME)
	if err != nil {
		t.Fatalf("failed to open (probably you need to set PGUSER, PGHOST, PGPORT or PGSSLMODE: %v", err)
	}
}

func TestMigrateNumberAtStart(t *testing.T) {
	m, err := NewPostgresMigrator(DB_NAME)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer m.Close()
	if err := m.DestroyMigrationRecords(); err != nil {
		t.Fatalf("Unable to clear migration records for test: %v", err)
	}
	record, err := m.CurrentMigrationNumber()
	if err != nil {
		t.Fatalf("Unable to fetch migration number: %v", err)
	}
	if record != 0 {
		t.Errorf("wrong migration found, expected 0 got %d", record)
	}
}

func checkMigrationApplication(t *testing.T, m Migrator, performed int, after int, fn func() (int, error)) {
	p, err := fn()
	if err != nil {
		t.Fatalf("failed to perform migration: %v", err)
	}
	if p != performed {
		t.Errorf("expected to have done %d migrations, but did %d", performed, p)
	}
	curr, err := m.CurrentMigrationNumber()
	if err != nil {
		t.Fatalf("could not check the current migration number: %v", err)
	}
	if curr != after {
		t.Errorf("expected to be at migration %d but really at %d", after, curr)
	}
}

func TestMigrateNumberUpDown(t *testing.T) {
	m, err := NewPostgresMigrator(DB_NAME)
	if err != nil {
		t.Fatalf("%v", err)
	}

	//after the test, just get rid of the things we tested
	defer func() {
		m.(*postgresMigrator).db.Exec("DROP TABLE foobar")
		m.(*postgresMigrator).db.Exec("DROP TABLE migrations")
		m.Close()
	}()

	if err := m.DestroyMigrationRecords(); err != nil {
		t.Fatalf("unable to clear the migration records: %v", err)
	}

	checkMigrationApplication(t, m, 2, 2, func() (int, error) {
		return m.Up(migs.Up)
	})

	//hacky way to simulate some data
	db := m.(*postgresMigrator).db
	if _, err = db.Exec("INSERT INTO foobar(i,t) VALUES (2,'blah'), (42,'answer')"); err != nil {
		t.Fatalf("error during insert: %v", err)
	}

	//down 1 migration
	checkMigrationApplication(t, m, 1, 1, func() (int, error) {
		return m.DownTo(1, migs.Down)
	})

	//check the values got converted
	var s1, s2 string
	if err = db.QueryRow("SELECT s FROM foobar WHERE i = 2").Scan(&s1); err != nil {
		t.Fatalf("error doing select of data converted (1): %v", err)
	}
	if s1 != "blah" {
		t.Errorf("expected blah but got %s", s1)
	}
	if err = db.QueryRow("SELECT s FROM foobar WHERE i = 42").Scan(&s2); err != nil {
		t.Fatalf("error doing select of data converted (2): %v", err)
	}
	if s2 != "answer" {
		t.Errorf("expected answer but got %s", s2)
	}

	//run last migration
	checkMigrationApplication(t, m, 1, 0, func() (int, error) {
		return m.DownTo(0, migs.Down)
	})

}
