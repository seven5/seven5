package seven5

import (
	"exp/sql"
	"fmt"
	sqlite3 "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path/filepath"
	"strings"
	"mongrel2"
)

type ProjectConfig struct {
	Path   string
	Name   string
	Logger *log.Logger
}

const (
	LOGDIR        = "log"
	NO_SUCH_TABLE = "no such table"
)

/*
handler_test = Handler(	send_spec='tcp://127.0.0.1:10070',
                       	send_ident='34f9ceee-cd52-4b7f-b197-88bf2f0ec378',
                       	recv_spec='tcp://127.0.0.1:10071',
			recv_ident='') 


main = Server(
    uuid="f400bf85-4538-4f7a-8908-67e313d515c2",
    access_log="/logs/access.log",
    error_log="/logs/error.log",
    chroot="./",
    default_host="localhost",
    name="test",
    pid_file="/run/mongrel2.pid",
    port=6767,
    hosts = [
        Host(name="localhost", routes={
            '/tests/': Dir(base='tests/', index_file='index.html',default_ctype='text/plain')
	    '/handlertest': handler_test
        })
    ]
)

servers = [main]
*/

func LocateProject(projectName string) (string, string) {
	_ = new(sqlite3.SQLiteDriver)
	cwd, _ := os.Getwd()
	arch := os.Getenv("GOARCH")
	localos := os.Getenv("GOOS")

	//this resets the cwd to the top of your project, assuming it exists
	//by assuming you are running in eclipse 
	eclipseBinDir := fmt.Sprintf("%s_%s", localos, arch)
	d, f := filepath.Split(cwd)
	if eclipseBinDir != "-" && eclipseBinDir == f {
		d = filepath.Clean(d)
		projectDir, b := filepath.Split(d)
		if b == "bin" {
			//possibly eclipse, switch to project root and look for
			//the project area
			pkg := filepath.Join(projectDir, "src", "pkg")
			info, _ := os.Stat(pkg)
			if info != nil && info.IsDirectory() {
				//probably running in eclipse, check the path to proj
				projectDir := filepath.Join(pkg, projectName)
				info, _ = os.Stat(projectDir)
				if info != nil && info.IsDirectory() {
					return projectDir, cwd //in eclipse
				}
			}
		}
	}
	//maybe you are in the project dir?
	foundProject := true
	candidate := cwd
	parts := strings.Split(projectName, string(filepath.Separator))
	for i := len(parts) - 1; i >= 0; i-- {
		child := parts[i]

		parent, kid := filepath.Split(candidate)
		if kid != child {
			foundProject = false
		}
		candidate = filepath.Clean(parent)
	}
	//did we walk up, checking package structure?	
	if foundProject {
		return cwd, cwd
	}

	//try the root of the big tarball
	guess := filepath.Join(cwd, projectName)
	info, _ := os.Stat(guess)
	if info != nil && info.IsDirectory() {
		return guess, cwd
	}

	return "", cwd
}

func VerifyProjectLayout(projectPath string) string {

	for _, dir := range []string{"handler", "rest", LOGDIR, "run", "static", "dynamic"} {
		if s, _ := os.Stat(filepath.Join(projectPath, dir)); s == nil || !s.IsDirectory() {
			return fmt.Sprintf("Unable to find %s\n", filepath.Join(projectPath, dir))
		}
	}
	return ""
}

func CreateLogger(projectPath string) (*log.Logger, string, error) {

	path := filepath.Join(projectPath, LOGDIR, "seven5.log")
	file, err := os.Create(path)
	if err != nil {
		return nil, "", err
	}

	result := log.New(file, "", log.LstdFlags|log.Lshortfile)
	return result, path, nil
}

func NewProjectConfig(path string, l *log.Logger) *ProjectConfig {
	result := new(ProjectConfig)
	result.Path = path
	_, n := filepath.Split(path)
	result.Name = n
	result.Logger = l
	return result
}

func ClearTestDB(config *ProjectConfig) error {

	db_path := filepath.Join(config.Path, fmt.Sprintf("%s_test.sql", config.Name))

	db, err := sql.Open("sqlite3", db_path)
	if err != nil {
		return err
	}

	config.Logger.Printf("created/found mongrel2 configuration db at path: %s", db_path)

	destroy := "delete from %s"

	tname := "handler"
	sql := fmt.Sprintf(destroy, tname)
	_, err = db.Exec(sql)
	if err != nil {
		if !strings.HasPrefix(err.Error(), NO_SUCH_TABLE) {
			config.Logger.Printf("error clearing table %s:%s", tname, err)
			return err
		} else {
			config.Logger.Printf("table \"%s\" not present, creating tables for mongrel2 configuration", tname)
		}
		//tables do not exist, create from scratch
		_, err = db.Exec(TABLEDEFS_SQL)
		if err!=nil {
			config.Logger.Printf("unable to create tables in mongrel2 config:%s",err.Error())
			return err
		}
	} else {
		//this is the case where the tables exist
		for _, tbl := range []string{"host", "log", "mimetype", "proxy", "route", "server", "setting", "statistic"} {
			sql := fmt.Sprintf(destroy, tname)
			_, err = db.Exec(sql)
			if err != nil {
				config.Logger.Printf("error clearing table %s:%s", tbl, err)
				return err
			}
		}
		config.Logger.Printf("cleared all table table from mongrel2 configuration")
	}

	return nil
}

func DiscoverHandlers(config *ProjectConfig) ([]*mongrel2.HandlerAddr,error) {
	return nil,nil
}


func generateHandlerConfig() {
}

func generate() {
}

func createDBTablesForMongrel2() {
}

const TABLEDEFS_SQL = `
CREATE TABLE handler (id INTEGER PRIMARY KEY,
    send_spec TEXT, 
    send_ident TEXT,
    recv_spec TEXT,
    recv_ident TEXT,
   raw_payload INTEGER DEFAULT 0,
   protocol TEXT DEFAULT 'json');
CREATE TABLE host (id INTEGER PRIMARY KEY, 
    server_id INTEGER,
    maintenance BOOLEAN DEFAULT 0,
    name TEXT,
    matching TEXT);
CREATE TABLE log(id INTEGER PRIMARY KEY,
    who TEXT,
    what TEXT,
    location TEXT,
    happened_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    how TEXT,
    why TEXT);
CREATE TABLE mimetype (id INTEGER PRIMARY KEY, mimetype TEXT, extension TEXT);
CREATE TABLE proxy (id INTEGER PRIMARY KEY,
    addr TEXT,
    port INTEGER);
CREATE TABLE route (id INTEGER PRIMARY KEY,
    path TEXT,
    reversed BOOLEAN DEFAULT 0,
    host_id INTEGER,
    target_id INTEGER,
    target_type TEXT);
CREATE TABLE server (id INTEGER PRIMARY KEY,
    uuid TEXT,
    access_log TEXT,
    error_log TEXT,
    chroot TEXT DEFAULT '/var/www',
    pid_file TEXT,
    default_host TEXT,
    name TEXT DEFAULT '',
    bind_addr TEXT DEFAULT "0.0.0.0",
    port INTEGER,
    use_ssl INTEGER default 0);
CREATE TABLE setting (id INTEGER PRIMARY KEY, key TEXT, value TEXT);
CREATE TABLE statistic (id SERIAL, 
    other_type TEXT,
    other_id INTEGER,
    name text,
    sum REAL,
    sumsq REAL,
    n INTEGER,
    min REAL,
    max REAL,
    mean REAL,
    sd REAL,
    primary key (other_type, other_id, name));
`
