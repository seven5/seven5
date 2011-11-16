//Web microframework for go.  See http://seven5.github.com/seven5
package seven5

import (
	"errors"
	"exp/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3" //force linkage,without naming the library
	"log"
	"mongrel2"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//ProjectConfig represents the configuration of a Seven5 web app.  This 
//structure's fields are discovered by the infrastructure, it is not specified
//explicitly.
type ProjectConfig struct {
	Path     string
	Name     string
	Logger   *log.Logger
	Handler  []*mongrel2.HandlerAddr
	ServerId string
}

const (
	TEST_SERVER_ID = "f548a1be-3c3c-4f0d-91bd-44edf672af31"

	LOGDIR   = "log"
	MONGREL2 = "mongrel2"
	OUR_LOG  = "seven5.log"

	LOCALHOST  = "localhost"
	ACCESS_LOG = "/" + LOGDIR + "/access.log"
	ERROR_LOG  = "/" + LOGDIR + "/error.log"
	PID_FILE   = RUN + "/mongrel2.pid"
	TEST_PORT  = 6767
	RUN        = "/run"
	CONTROL    = "control"

	STATIC = "static/" //must have trailing slash

	NO_SUCH_TABLE    = "no such table"
	HANDLER_SUFFIX   = "_handler.go"
	HANDLER_INSERT   = `insert into handler(send_spec,send_ident,recv_spec,recv_ident) values("%s","%s","%s","");`
	HOST_INSERT      = `insert into host(server_id,name,matching) values(last_insert_rowid(),"%s","%s");`
	SERVER_INSERT    = `insert into server(uuid,access_log,error_log,pid_file,chroot,default_host,name,port) values("%s","%s","%s","%s","%s","%s","%s","%d");`
	MIME_INSERT      = `insert into mimetype(mimetype,extension) values("%s","%s");`
	DIRECTORY_INSERT = `insert into directory(base,index_file,default_ctype) values("%s","index.html","text/plain");`
	ROUTE_INSERT     = `insert into route(path,host_id,target_id,target_type) values("%s",%s, %s, "%s");`
	LOG_INSERT       = `insert into log(who,what,location,how,why) values("user","load", "localhost", "%s","webapp_start");`
)

//VerifyProjectLayout checks that the directory structure that is expected for
//a correct Seven5 application is present.  
func VerifyProjectLayout(projectPath string) string {

	for _, dir := range []string{LOGDIR, RUN, STATIC} {
		candidate := filepath.Join(projectPath, MONGREL2, dir)
		if s, _ := os.Stat(filepath.Clean(candidate)); s == nil || !s.IsDirectory() {
			return fmt.Sprintf("Unable to find directory %s", candidate)
		}
	}
	return ""
}

//CreateLogger builds a logger that is connected to the "standard" place in 
//a Seven5 application (mongrel2/log/seven5.log)
func CreateLogger(projectPath string) (*log.Logger, string, error) {

	path := filepath.Join(projectPath, MONGREL2, LOGDIR, OUR_LOG)
	file, err := os.Create(path)
	if err != nil {
		return nil, "", err
	}

	result := log.New(file, "", log.LstdFlags|log.Lshortfile)
	return result, path, nil
}

//Create a new ProjectConfig object and return a pointer to it.  This method
//fills in some fields that are already known such as the path, name, and 
//logger.
func NewProjectConfig(path string, l *log.Logger) (*ProjectConfig, error) {
	result := new(ProjectConfig)
	p, err := filepath.Abs(path)
	result.Path = p

	if err != nil {
		return nil, err
	}
	_, n := filepath.Split(filepath.Clean(path))
	result.Name = n
	result.Logger = l
	return result, nil
}

//ClearTestDB opens the sqlite3 database for mongrel3 configuration and insures
//that the tables are present and empty.  If the tables are not present, this
//function creates them.
func ClearTestDB(config *ProjectConfig) error {

	db, err := sql.Open("sqlite3", db_path(config))
	if err != nil {
		return err
	}

	config.Logger.Printf("created/found mongrel2 configuration db at path: %s", db_path(config))

	//destroy all tables, create from scratch
	for _, create := range TABLEDEFS_SQL {
		_, err = db.Exec(create)
		if err != nil {
			config.Logger.Printf("unable to create tables in mongrel2 config:%s", err.Error())
			return err
		}
	}

	config.Logger.Printf("dropped all mongrel config tables and re-created them")
	config.ServerId = TEST_SERVER_ID

	return nil
}

//DiscoverHandlers looks for handlers in the standard place in a Seven5 project
//(project_name/*_handler.go) and assigns each one a mongrel2 address.  If
//there is no error, it puts the assigned mongrel2 addresses (type is
//mongrel2.HandlerAddr) into the ProjectConfig sruct.
func DiscoverHandlers(config *ProjectConfig) error {

	dir, err := os.Open(config.Path)
	if err != nil {
		return err
	}
	children, err := dir.Readdirnames(0)
	if err != nil {
		return err
	}
	address := []*mongrel2.HandlerAddr{}
	for _, child := range children {
		if !strings.HasSuffix(child, HANDLER_SUFFIX) {
			continue
		}
		info, _ := os.Stat(filepath.Join(config.Path, child))
		if info == nil || !info.IsRegular() {
			continue
		}
		end := strings.Index(child, HANDLER_SUFFIX)
		name := child[0:end]
		a, err := mongrel2.GetHandlerAddress(name)
		if err != nil {
			return nil
		}
		config.Logger.Printf("assigned raw handler %s unique id %s\n", name, a.UUID)
		address = append(address, a)
	}
	config.Handler = address
	return nil
}

//GenerateMongrelConfig all the necessary configuration for the mongrel2 server into the 
//database in the standard location.  It sets all the variables of the mongrel2 instance
//to their default values.
func GenerateMongrel2Config(config *ProjectConfig) error {

	db, err := sql.Open("sqlite3", db_path(config))
	if err != nil {
		return err
	}

	var sqlText string

	//note: these are ORDER DEPENDENT because of selects later than reference tables 
	//note: created earlier

	//SERVER
	chroot := filepath.Join(config.Path, MONGREL2)
	sqlText = fmt.Sprintf(SERVER_INSERT, config.ServerId, ACCESS_LOG, ERROR_LOG, PID_FILE, chroot, LOCALHOST, config.Name, TEST_PORT)
	_, err = db.Exec(sqlText)
	if err != nil {
		return err
	}
	config.Logger.Printf("=== server id %s ===\n", config.ServerId)
	config.Logger.Printf("inserted server into config:%s\n", sqlText)

	//HOST
	sqlText = fmt.Sprintf(HOST_INSERT, LOCALHOST, LOCALHOST)
	_, err = db.Exec(sqlText)
	if err != nil {
		return err
	}
	config.Logger.Printf("inserted host into config:%s\n", sqlText)

	//HANDLER
	for _, addr := range config.Handler {
		sqlText = fmt.Sprintf(HANDLER_INSERT, addr.PullSpec, addr.UUID, addr.PubSpec)
		_, err := db.Exec(sqlText)
		if err != nil {
			return err
		}
		config.Logger.Printf("inserted %s handler configuration:%s\n", addr.Name, sqlText)
	}

	

	//this is the query used to find the host that routes point at
	nestedHost := fmt.Sprintf(`(select id from host where name="%s")`, LOCALHOST)

	// ROUTE TO HANDLERS
	for _, addr := range config.Handler {
		nestedHandler := fmt.Sprintf(`(select id from handler where send_spec="%s")`, addr.PullSpec)
		sqlText = fmt.Sprintf(ROUTE_INSERT, "/"+addr.Name, nestedHost, nestedHandler, "handler")
		_, err = db.Exec(sqlText)
		if err != nil {
			return err
		}
		config.Logger.Printf("inserted handler %s into routes:%s\n", addr.Name, sqlText)
		//config.Logger.Printf("\tnested host:%s\n",nestedHost)
		//config.Logger.Printf("\tnested handler:%s\n",nestedHandler)
	}

	//static content
	sqlText = fmt.Sprintf(DIRECTORY_INSERT, STATIC)
	_, err = db.Exec(sqlText)
	if err != nil {
		return err
	}
	config.Logger.Printf("inserted directory into config:%s\n", DIRECTORY_INSERT)

	// ROUTE TO STATIC CONTENT
	staticDirectory := fmt.Sprintf(`(select id from directory where base="%s")`, STATIC)
	sqlText = fmt.Sprintf(ROUTE_INSERT, "/", nestedHost, staticDirectory, "dir")

	_, err = db.Exec(sqlText)
	if err != nil {
		return err
	}
	config.Logger.Printf("inserted static content route into config:%s\n", sqlText)
	//config.Logger.Printf("\tnested host:%s\n",nestedHost)
	//config.Logger.Printf("\tnested directory:%s\n",staticDirectory)

	for _, pair := range MIME_TYPE {
		sqlText = fmt.Sprintf(MIME_INSERT, pair[0], pair[1])
		_, err = db.Exec(sqlText)
		if err != nil {
			return err
		}
	}
	config.Logger.Printf("inserted %d mime types in mongrel2 configuration\n", len(MIME_TYPE))

	sqlText = fmt.Sprintf(LOG_INSERT, db_path(config))
	_, err = db.Exec(sqlText)
	if err != nil {
		return err
	}
	config.Logger.Printf("inserted log entry into config:%s\n", sqlText)

	return nil
}

//Bootstrap should be called by projects that use the Seven5 infrastructure to
//parse the command line arguments and to discover the project's structure.
//This function returns null if the project config is not standard.  
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
		return nil
	}

	logger, path, err := CreateLogger(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create logger:%s\n", err)
		return nil
	}

	fmt.Printf("Seven5 is logging to %s\n", path)

	config, err := NewProjectConfig(projectDir, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to compute project configuration (problem with path to configuration db?):%s\n", err.Error())
		return nil
	}

	config.Logger.Printf("Starting to run with project %s at %s", config.Name, config.Path)

	err = ClearTestDB(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error clearing the test db configuration:%s\n", err.Error())
		return nil
	}

	err = DiscoverHandlers(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to discover mongrel2 handlers:%s\n", err.Error())
		return nil
	}

	err = GenerateMongrel2Config(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to write mongrel2 config:%s\n", err.Error())
		return nil
	}

	err = runMongrel(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to start/reset mongrel2:%s\n", err.Error())
		return nil
	}

	return config
}

func startMongrel(config *ProjectConfig) error {

	mongrelPath, err := exec.LookPath("mongrel2")
	if err != nil {
		config.Logger.Printf("unable to find mongrel2 in path! you should probably change your PATH environment variable")
		return err
	}

	config.Logger.Printf("using mongrel2 binary at %s", mongrelPath)

	cmd := exec.Command(mongrelPath, db_path(config), config.ServerId)
	cmd.Dir = filepath.Join(config.Path, "mongrel2")

	err = cmd.Start()
	if err != nil {
		return err
	}

	config.Logger.Printf("changed working directory to %s and started mongrel2 (pid=%d)\n", cmd.Dir, cmd.Process.Pid)

	return nil
}

func runMongrel(config *ProjectConfig) error {

	socketPath := filepath.Join(config.Path, MONGREL2, RUN, CONTROL)

	addr, err := net.ResolveUnixAddr("unix", socketPath)
	if err != nil {
		return err
	}

	unixConn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		config.Logger.Printf("Unable to connect to mongrel2 via %s (%s): will try to start mongrel2\n", socketPath, err)
		return startMongrel(config)
	}

	config.Logger.Printf("connected to mongrel2 via %s... restarting\n", socketPath)

	reload := "reload\n"
	n, err := unixConn.Write([]byte(reload))
	if err != nil {
		return err
	}
	if n != len(reload) {
		return errors.New("didn't write the correct number of bytes on unix socket for reload!")
	}

	return nil

}

//dumpBadProjectLayout prints out a helpful error message to stderr so the 
//developer has some hope of figuring out what is wrong with they project 
//structure.
func dumpBadProjectLayout(projectDir, err string) {
	fmt.Fprintf(os.Stderr, "%s does not have the standard seven5 project structure!\n", projectDir)
	fmt.Fprintf(os.Stderr, "\t(%s)\n", err)
	fmt.Fprintf(os.Stderr, "\nfor project structure details, see http://seven5.github.com/seven5/project_layout.html\n\n")
}

//Create a path to the SQLite db, given a project structure
func db_path(config *ProjectConfig) string {
	return filepath.Join(config.Path, fmt.Sprintf("%s_test.sqlite", config.Name))

}

var TABLEDEFS_SQL = []string{
	`DROP TABLE IF EXISTS server;`,
	`DROP TABLE IF EXISTS host;`,
	`DROP TABLE IF EXISTS handler;`,
	`DROP TABLE IF EXISTS proxy;`,
	`DROP TABLE IF EXISTS route;`,
	`DROP TABLE IF EXISTS statistic;`,
	`DROP TABLE IF EXISTS mimetype;`,
	`DROP TABLE IF EXISTS setting;`,
	`DROP TABLE IF EXISTS directory;`,
	`CREATE TABLE handler (id INTEGER PRIMARY KEY,
    send_spec TEXT, 
    send_ident TEXT,
    recv_spec TEXT,
    recv_ident TEXT,
   raw_payload INTEGER DEFAULT 0,
   protocol TEXT DEFAULT 'json');`,
	`CREATE TABLE directory (id INTEGER PRIMARY KEY,   base TEXT,   index_file TEXT,  	default_ctype TEXT,   cache_ttl INTEGER DEFAULT 0);`,
	`CREATE TABLE host (id INTEGER PRIMARY KEY, 
    server_id INTEGER,
    maintenance BOOLEAN DEFAULT 0,
    name TEXT,
    matching TEXT);`,
	`CREATE TABLE IF NOT EXISTS log(id INTEGER PRIMARY KEY,
    who TEXT,
    what TEXT,
    location TEXT,
    happened_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    how TEXT,
    why TEXT);`,
	`CREATE TABLE mimetype (id INTEGER PRIMARY KEY, mimetype TEXT, extension TEXT);`,
	`CREATE TABLE proxy (id INTEGER PRIMARY KEY,
    addr TEXT,
    port INTEGER);`,
	`CREATE TABLE route (id INTEGER PRIMARY KEY,
    path TEXT,
    reversed BOOLEAN DEFAULT 0,
    host_id INTEGER,
    target_id INTEGER,
    target_type TEXT);`,
	`CREATE TABLE server (id INTEGER PRIMARY KEY,
    uuid TEXT,
    access_log TEXT,
    error_log TEXT,
    chroot TEXT DEFAULT '/var/www',
    pid_file TEXT,
    default_host TEXT,
    name TEXT DEFAULT '',
    bind_addr TEXT DEFAULT "0.0.0.0",
    port INTEGER,
    use_ssl INTEGER default 0);`,
	`CREATE TABLE setting (id INTEGER PRIMARY KEY, key TEXT, value TEXT);`,
	`CREATE TABLE statistic (id SERIAL, 
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
`,
}
