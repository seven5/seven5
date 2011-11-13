package seven5

import (
	"exp/sql"
	"flag"
	"fmt"
	sqlite3 "github.com/mattn/go-sqlite3"
	"log"
	"mongrel2"
	"os"
	"path/filepath"
	"strings"
)

type ProjectConfig struct {
	Path   string
	Name   string
	Logger *log.Logger
}

const (
	LOGDIR        = "log"
	MONGREL2        = "mongrel2"
	NO_SUCH_TABLE = "no such table"
)
// Try to look at the layout of the filesystem and figure out where the 
// eclipse source code might be.  
func LocateProjectInEclipse() string {
	_ = new(sqlite3.SQLiteDriver)
	cwd, _ := os.Getwd()
	arch := os.Getenv("GOARCH")
	localos := os.Getenv("GOOS")

	//this resets the cwd to the top of your project, assuming it exists
	//by assuming you are running in eclipse 
	eclipseBinDir := fmt.Sprintf("%s_%s", localos, arch)
	d, f := filepath.Split(cwd)
	if eclipseBinDir != "_" && eclipseBinDir == f {
		d = filepath.Clean(d)
		projectDir, b := filepath.Split(d)
		if b == "bin" {
			//possibly eclipse, switch to project root and look for
			//the project area
			pkg := filepath.Join(projectDir, "src", "pkg")
			info, _ := os.Stat(pkg)
			if info != nil && info.IsDirectory() {
				dir, err := os.Open(pkg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "unable to open directory '%s'\n", pkg)
				} else {
					name, err := dir.Readdirnames(0) // get all names
					if err != nil {
						fmt.Fprintf(os.Stderr, "cannot read children of directory '%s'!\n", pkg)
					} else {
						if len(name) != 1 {
							fmt.Fprintf(os.Stderr, "directory '%s' has '%d' children, can't decide which one to use!\n", len(name))
						} else {
							//probably running in eclipse, check the path to proj
							projectDir := filepath.Join(pkg, name[0])
							info, _ = os.Stat(projectDir)
							if info != nil && info.IsDirectory() {
								return projectDir
							}
						}
					}
				}
			}
		}
	}
	return ""
}

func VerifyProjectLayout(projectPath string) string {

	for _, dir := range []string{LOGDIR, "run", "static"} {
		candidate:=filepath.Join(projectPath, MONGREL2, dir)
		if s, _ := os.Stat(candidate); s == nil || !s.IsDirectory() {
			return fmt.Sprintf("Unable to find directory %s", candidate)
		}
	}
	return ""
}

func CreateLogger(projectPath string) (*log.Logger, string, error) {

	path := filepath.Join(projectPath, MONGREL2, LOGDIR, "seven5.log")
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

	db_path := filepath.Join(config.Path, fmt.Sprintf("%s_test.sqlite", config.Name))

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
		if err != nil {
			config.Logger.Printf("unable to create tables in mongrel2 config:%s", err.Error())
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

func DiscoverHandlers(config *ProjectConfig) ([]*mongrel2.HandlerAddr, error) {
	return nil, nil
}

func generateHandlerConfig() {
}

func Bootstrap() *ProjectConfig {

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to determine current directory: %s\n", err)
		return nil
	}

	flagSet := flag.NewFlagSet("seven5", flag.ExitOnError)
	flagSet.Parse(os.Args)

	projectDir := cwd

	if flagSet.NArg() == 1 {
		fmt.Fprintf(os.Stderr, "no project name/path specified, hoping '%s' is ok.\n", cwd)
	} else {
		projectDir = flagSet.Arg(1)
	}

	if err := VerifyProjectLayout(projectDir); err != "" {
		dumpBadProjectLayout(projectDir, err)
		if eclipse := LocateProjectInEclipse(); eclipse != "" {
			fmt.Fprintf(os.Stderr, "checking for possible eclipse project at '%s'\n",eclipse)
			if err := VerifyProjectLayout(eclipse); err != "" {
				dumpBadProjectLayout(eclipse, err)
				return nil
			} else {
				projectDir=eclipse //success! found eclipse project!
			}
		} else {
			return nil
		}
	}

	logger, path, err := CreateLogger(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create logger:%s\n", err)
		return nil
	}

	fmt.Printf("Seven5 is logging to %s\n", path)

	config := NewProjectConfig(projectDir, logger)
	config.Logger.Printf("Starting to run with project %s at %s", config.Name, config.Path)

	err = ClearTestDB(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error clearing the test db configuration:%s\n", err.Error())
		return nil
	}

	_, err = DiscoverHandlers(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to discover mongrel2 handlers:%s\n", err.Error())
		return nil
	}

	return config
}

func generate() {
}

func dumpBadProjectLayout(projectDir, err string) {
	fmt.Fprintf(os.Stderr, "%s does not have the standard seven5 project structure!\n", projectDir)
	fmt.Fprintf(os.Stderr, "\t(%s)\n", err)
	fmt.Fprintf(os.Stderr, "\nfor project structure details, see http://seven5.github.com/seven5/project_layout.html\n\n")
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
