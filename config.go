//Web microframework for go.  See http://seven5.github.com/seven5
package seven5

import (
	"errors"
	"exp/sql"
	"flag"
	"fmt"
	"github.com/alecthomas/gozmq"
	_ "github.com/mattn/go-sqlite3" //force linkage,without naming the library
	"log"
	"mongrel2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

//ProjectConfig represents the configuration of a Seven5 web app.  This 
//structure's fields are discovered by the infrastructure, it is not specified
//explicitly.
type ProjectConfig struct {
	Path     string
	Name     string
	Logger   *log.Logger
	Addr     []*mongrel2.HandlerSpec
	Service  []string
	ServerId string
	Db	     *sql.DB
}

const (
	TEST_SERVER_ID = "f548a1be-3c3c-4f0d-91bd-44edf672af31"

	LOGDIR   = "log"
	MONGREL2 = "mongrel2"
	OUR_LOG  = "seven5.log"

	WEBAPP_START_DIR = "webapp_start"

	LOCALHOST  = "localhost"
	ACCESS_LOG = "/" + LOGDIR + "/access.log"
	ERROR_LOG  = "/" + LOGDIR + "/error.log"
	PID_FILE   = RUN + "/mongrel2.pid"
	TEST_PORT  = 6767
	RUN        = "/run"
	CONTROL    = "control"

	TIME_CMD                = "10:4:time,0:}]"
	RELOAD_CMD              = "12:6:reload,0:}]"
	TWO_SECS_IN_MICROS      = 2000000
	HUNDRED_MSECS_IN_MICROS = 100000

	STATIC = "static/" //must have trailing slash

	NO_SUCH_TABLE = "no such table"

	HANDLER_SUFFIX = "_rawhttp.go"
	SERVICE_SUFFIX = "_jsonservice.go"

	HANDLER_OR_SERVICE_INSERT = `insert into handler(send_spec,send_ident,recv_spec,recv_ident) values("%s","%s","%s","");`
	HOST_INSERT               = `insert into host(server_id,name,matching) values(last_insert_rowid(),"%s","%s");`
	SERVER_INSERT             = `insert into server(uuid,access_log,error_log,pid_file,chroot,default_host,name,port) values("%s","%s","%s","%s","%s","%s","%s","%d");`
	MIME_INSERT               = `insert into mimetype(mimetype,extension) values("%s","%s");`
	DIRECTORY_INSERT          = `insert into directory(base,index_file,default_ctype) values("%s","index.html","text/plain");`
	ROUTE_INSERT              = `insert into route(path,host_id,target_id,target_type) values("%s",%s, %s, "%s");`
	LOG_INSERT                = `insert into log(who,what,location,how,why) values("user","load", "localhost", "%s","webapp_start");`
)

//VerifyProjectLayout checks that the directory structure that is expected for
//a correct Seven5 application is present.  
func VerifyProjectLayout(projectPath string) string {
	for _, dir := range []string{LOGDIR, RUN, STATIC} {
		candidate := filepath.Join(projectPath, MONGREL2, dir)
		clean := filepath.Clean(candidate)
		if s, _ := os.Stat(clean); s == nil || !s.IsDirectory() {
			return fmt.Sprintf("Unable to find directory %s (%v)", clean, s == nil)
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

	config.Logger.Printf("[CONFIG] created/found mongrel2 configuration db at path: %s", db_path(config))

	//destroy all tables, create from scratch
	for _, create := range TABLEDEFS_SQL {
		_, err = db.Exec(create)
		if err != nil {
			config.Logger.Printf("[ERROR!] unable to create tables in mongrel2 config:%s", err.Error())
			return err
		}
	}

	config.ServerId = TEST_SERVER_ID

	return nil
}

//DiscoverHandlers looks for handlers in the standard place in a Seven5 project
//(project_name/*_handler.go) and assigns each one a mongrel2 address. 
//It also looks for project_name/*_service.go and assigns these a mongrel address. If
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
	address := []*mongrel2.HandlerSpec{}
	service := []string{}
	for _, child := range children {
		if !strings.HasSuffix(child, HANDLER_SUFFIX) && !strings.HasSuffix(child, SERVICE_SUFFIX) {
			continue
		}
		info, _ := os.Stat(filepath.Join(config.Path, child))
		if info == nil || !info.IsRegular() {
			continue
		}
		var name string
		if strings.HasSuffix(child, HANDLER_SUFFIX) {
			end := strings.Index(child, HANDLER_SUFFIX)
			name = child[0:end]
			config.Logger.Printf("[DISCOVER] found raw handler %s\n",name)
		} else {
			end := strings.Index(child, SERVICE_SUFFIX)
			name = child[0:end]
			config.Logger.Printf("[DISCOVER] found service %s\n",name)
			service = append(service, name)
		}
		a, err := mongrel2.GetHandlerSpec(name)
		if err != nil {
			return nil
		}
		address = append(address, a)
	}
	config.Addr = address
	config.Service = service
	return nil
}

//GenerateServerConfig writes information about the server (in mongrel2 terms) name and what host it is
//working on behalf of.  It also configures information about the log file placement
//and other such variables.  The optional arguments are
//accessLog string, errorLog string, pidFile string, chroot string. The "host" portion of the
//mongrel2 configuration is made to be exactly one host whose name is default host.
func GenerateServerHostConfig(config *ProjectConfig,defaultHost string, port int,  opt ...string) error {

	if len(opt)!=4 && len(opt)!=0 {
		return errors.New("GenerateServerConfig: wrong number of arguments!")
	}
	var chroot, accessLog, errorLog, pidFile string
	if len(opt)==0 {
		chroot = filepath.Join(config.Path, MONGREL2)
		accessLog=ACCESS_LOG
		errorLog=ERROR_LOG
		pidFile=PID_FILE
	} else {
		accessLog=opt[0]
		errorLog=opt[1]
		pidFile=opt[2]
		chroot = opt[3]
	}
	sqlText := fmt.Sprintf(SERVER_INSERT, config.ServerId, accessLog, errorLog, pidFile, chroot, defaultHost, config.Name, port)
	r, err := config.Db.Exec(sqlText)
	if err != nil {
		return err
	}
	res, err := r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[MONGREL2 SQL] inserted server into config:%s (%d row)\n", sqlText, res)

	//HOST: must run IMMEDIATELY after the server is inserted because sql is dependent on order
	sqlText = fmt.Sprintf(HOST_INSERT, defaultHost, defaultHost)
	r, err = config.Db.Exec(sqlText)
	if err != nil {
		return err
	}
	res, err = r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[MONGREL2 SQL] inserted host into config:%s (%d row)\n", sqlText, res)
	return nil
}

//GenerateHandlerAddressAndRouteConfig creates the necessary mongrel2 configuration for a handler
//with a given name. It creates an address for the handler, an ID for it and writes those
//into the mongrel2 configuration.  Then it creates a mongrel2 route for the handler to 
//be contacted on.  The route is @[name] if the handler is a JSON service, otherwise it is
//bound into the HTTP space at /[name]
func GenerateHandlerAddressAndRouteConfig(config *ProjectConfig, host string, handler Named, isJson bool) error {
	addr,err:=mongrel2.GetHandlerSpec(handler.Name())
	if err!=nil {
		return err
	}
	sqlText := fmt.Sprintf(HANDLER_OR_SERVICE_INSERT, addr.PullSpec, addr.Identity, addr.PubSpec)
	r, err := config.Db.Exec(sqlText)
	if err != nil {
		return err
	}
	res, err := r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[MONGREL2 SQL] inserted %s handler configuration:%s (%d row)\n", addr.Name, sqlText, res)

	//this is the query used to find the host that routes point at
	nestedHost := fmt.Sprintf(`(select id from host where name="%s")`, host)

	// ROUTE TO HANDLER
	nestedHandler := fmt.Sprintf(`(select id from handler where send_spec="%s")`, addr.PullSpec)
	pathPrefix := "/"
	if isJson {
		pathPrefix="@"
	}
	sqlText = fmt.Sprintf(ROUTE_INSERT, pathPrefix+addr.Name, nestedHost, nestedHandler, "handler")
	r, err = config.Db.Exec(sqlText)
	if err != nil {
		return err
	}
	res, err = r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[MONGREL2 SQL] inserted handler %s into routes:%s (%d rows affected)\n", addr.Name, sqlText, res)
	return nil
} 

//GenerateStaticContentConfigcreates the necessary mongrel2 configuration for a directory
//a 'directory' and a 'route' to that directory.
func GenerateStaticContentConfig(config *ProjectConfig, host string, path string) error {
	//this is the query used to find the host that routes point at
	nestedHost := fmt.Sprintf(`(select id from host where name="%s")`, host)
	
	
	dirText := fmt.Sprintf(DIRECTORY_INSERT, path)
	r, err := config.Db.Exec(dirText)
	if err != nil {
		return err
	}
	res, err := r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[MONGREL2 SQL] inserted directory into config:%s (%d row)\n", dirText, res)

	// ROUTE TO STATIC CONTENT
	staticDirectory := fmt.Sprintf(`(select id from directory where base="%s")`, STATIC)
	routeText := fmt.Sprintf(ROUTE_INSERT, "/static/", nestedHost, staticDirectory, "dir")

	r, err = config.Db.Exec(routeText)
	if err != nil {
		return err
	}
	res, err = r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[MONGREL2 SQL] inserted static content route into config:%s (%d row)\n", routeText, res)
	return nil
}

//GenerateMimeTypeConfig puts the mime-type table.
func GenerateMimeTypeConfig(config *ProjectConfig) error {
	rows, err := config.Db.Query("select count(*) from mimetype;")
	if err != nil {
		return err
	}

	var result int
	rows.Next()
	rows.Scan(&result)
	rows.Close()

	config.Logger.Printf("[MONGREL2 SQL] currenty %d items in mimetype table\n", result)

	if result == 0 {
		for _, pair := range MIME_TYPE {
			mimeText := fmt.Sprintf(MIME_INSERT, pair[0], pair[1])
			_, err = config.Db.Exec(mimeText)
			if err != nil {
				return err
			}
		}
		config.Logger.Printf("[MONGREL2 SQL] inserted %d mime types in mongrel2 configuration\n", len(MIME_TYPE))
	}
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
	r, err := db.Exec(sqlText)
	if err != nil {
		return err
	}
	res, err := r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[MONGREL2 SQL] inserted server into config:%s (%d row)\n", sqlText, res)

	//HOST
	sqlText = fmt.Sprintf(HOST_INSERT, LOCALHOST, LOCALHOST)
	r, err = db.Exec(sqlText)
	if err != nil {
		return err
	}
	res, err = r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[MONGREL2 SQL] inserted host into config:%s (%d row)\n", sqlText, res)

	//HANDLER
	for _, addr := range config.Addr {
		sqlText = fmt.Sprintf(HANDLER_OR_SERVICE_INSERT, addr.PullSpec, addr.Identity, addr.PubSpec)
		r, err = db.Exec(sqlText)
		if err != nil {
			return err
		}
		res, err = r.RowsAffected()
		if err != nil {
			return err
		}
		config.Logger.Printf("[MONGREL2 SQL] inserted %s handler configuration:%s (%d row)\n", addr.Name, sqlText, res)
	}

	//this is the query used to find the host that routes point at
	nestedHost := fmt.Sprintf(`(select id from host where name="%s")`, LOCALHOST)

	// ROUTE TO HANDLERS
	for _, addr := range config.Addr {
		nestedHandler := fmt.Sprintf(`(select id from handler where send_spec="%s")`, addr.PullSpec)
		pathPrefix := "/"
		//services use different insert ...
		for _, n := range config.Service {
			if n == addr.Name {
				pathPrefix = "@"
				break
			}
		}
		sqlText = fmt.Sprintf(ROUTE_INSERT, pathPrefix+addr.Name, nestedHost, nestedHandler, "handler")
		_, err = db.Exec(sqlText)
		if err != nil {
			return err
		}
		config.Logger.Printf("[MONGREL2 SQL] inserted handler %s into routes:%s\n", addr.Name, sqlText)
	}

	//static content
	dirText := fmt.Sprintf(DIRECTORY_INSERT, STATIC)
	r, err = db.Exec(dirText)
	if err != nil {
		return err
	}
	res, err = r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[MONGREL2 SQL] inserted directory into config:%s (%d row)\n", dirText, res)

	// ROUTE TO STATIC CONTENT
	staticDirectory := fmt.Sprintf(`(select id from directory where base="%s")`, STATIC)
	routeText := fmt.Sprintf(ROUTE_INSERT, "/static/", nestedHost, staticDirectory, "dir")

	r, err = db.Exec(routeText)
	if err != nil {
		return err
	}
	res, err = r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[MONGREL2 SQL] inserted static content route into config:%s (%d row)\n", routeText, res)

	rows, err := db.Query("select count(*) from mimetype;")
	if err != nil {
		return err
	}

	var result int
	rows.Next()
	rows.Scan(&result)
	rows.Close()

	config.Logger.Printf("[MONGREL2 SQL] currenty %d items in mimetype table\n", result)

	if result == 0 {
		for _, pair := range MIME_TYPE {
			mimeText := fmt.Sprintf(MIME_INSERT, pair[0], pair[1])
			_, err = db.Exec(mimeText)
			if err != nil {
				return err
			}
		}
		config.Logger.Printf("[MONGREL2 SQL] inserted %d mime types in mongrel2 configuration\n", len(MIME_TYPE))
	}

	/* USELESS
	sqlText = fmt.Sprintf(LOG_INSERT, db_path(config))
	_, err = db.Exec(sqlText)
	if err != nil {
		return err
	}
	config.Logger.Printf("inserted log entry into config:%s\n", sqlText)
	*/

	return nil
}

//Bootstrap should be called by projects that use the Seven5 infrastructure to
//parse the command line arguments and to discover the project's structure.
//This function returns null if the project config is not standard.  This is a 
//wrapper on BootstrapFromDir that gets the project directory from the command line args. 
func Bootstrap() (*ProjectConfig, gozmq.Context) {

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to determine current directory: %s\n", err)
		return nil, nil
	}

	flagSet := flag.NewFlagSet("seven5", flag.ExitOnError)
	flagSet.Parse(os.Args)

	projectDir := cwd

	if flagSet.NArg() == 1 {
		fmt.Fprintf(os.Stderr, "no project name/path specified, hoping '%s' is ok.\n", cwd)
	} else {
		projectDir = flagSet.Arg(1)
	}
	return BootstrapFromDir(projectDir)
}

//BootstrapFromDir does the heavy lifting to set up a project, given a directory to work
//with.  It returns a configuration (or nil on error) plus the Context object that it 
//created for use in this application.
func BootstrapFromDir(projectDir string) (*ProjectConfig, gozmq.Context) {
	if err := VerifyProjectLayout(projectDir); err != "" {
		dumpBadProjectLayout(projectDir, err)
		return nil, nil
	}

	logger, path, err := CreateLogger(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create logger:%s\n", err)
		return nil, nil
	}

	fmt.Printf("Seven5 is logging to %s\n", path)

	config, err := NewProjectConfig(projectDir, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to compute project configuration (problem with path to configuration db?):%s\n", err.Error())
		return nil, nil
	}

	config.Logger.Printf("---- PROJECT %s @ %s ----", config.Name, config.Path)

	err = ClearTestDB(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error clearing the test db configuration:%s\n", err.Error())
		return nil, nil
	}

	/*
	err = DiscoverHandlers(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to discover mongrel2 handlers:%s\n", err.Error())
		return nil, nil
	}
	
	err = GenerateMongrel2Config(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to write mongrel2 config:%s\n", err.Error())
		return nil, nil
	}
	*/
	
	ctx, err := gozmq.NewContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create zmq context :%s\n", err.Error())
		return nil, nil
	}

	err = runMongrel(config, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to start/reset mongrel2:%s\n", err.Error())
		return nil, nil
	}

	db, err := sql.Open("sqlite3", db_path(config))
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to open configuration database:%s\n", err.Error())
		return nil,nil
	}
	config.Db=db

	return config, ctx
}

func startMongrel(config *ProjectConfig, ctx gozmq.Context) error {

	mongrelPath, err := exec.LookPath("mongrel2")
	if err != nil {
		config.Logger.Printf("[ERROR!] unable to find mongrel2 in path! you should probably change your PATH environment variable")
		return err
	}

	config.Logger.Printf("[PROCESS] using mongrel2 binary at %s", mongrelPath)

	stdout, err := os.Create(filepath.Join(config.Path, MONGREL2, LOGDIR, "mongrel2.out.log"))
	if err != nil {
		return err
	}

	stderr, err := os.Create(filepath.Join(config.Path, MONGREL2, LOGDIR, "mongrel2.err.log"))
	if err != nil {
		return err
	}

	cmd := exec.Command(mongrelPath, db_path(config), config.ServerId)
	cmd.Dir = filepath.Join(config.Path, "mongrel2")

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	cmd.SysProcAttr = new(syscall.SysProcAttr)
	cmd.SysProcAttr.Setsid = true

	err = cmd.Start()
	if err != nil {
		return err
	}

	config.Logger.Printf("[PROCESS] changed working directory to %s and started mongrel2 (pid=%d)\n", cmd.Dir, cmd.Process.Pid)

	socketPath := filepath.Join(config.Path, MONGREL2, RUN, CONTROL)
	path := fmt.Sprintf("ipc://%s", socketPath)
	config.Logger.Printf("[ZMQ] using zmq connection to mongrel2 for sync with startup '%s'", path)
	response, err := SendMongrelControl(TIME_CMD, TWO_SECS_IN_MICROS, path, ctx)
	if err != nil {
		return err
	}
	config.Logger.Printf("[ZMQ] time command on startup succeeded, response was '%s'", response)

	return nil
}

func runMongrel(config *ProjectConfig, ctx gozmq.Context) error {

	socketPath := filepath.Join(config.Path, MONGREL2, RUN, CONTROL)
	path := fmt.Sprintf("ipc://%s", socketPath)
	config.Logger.Printf("[ZMQ] zmq connection to mongrel2 is '%s'", path)

	result, err := SendMongrelControl(RELOAD_CMD, HUNDRED_MSECS_IN_MICROS, path, ctx)
	if err != nil {
		return err
	}
	if result == "" {
		config.Logger.Printf("[ZMQ] did not get any response on mongrel control connection... will try to start mongrel")
		return startMongrel(config, ctx)
	}
	config.Logger.Printf("[ZMQ] reload successful: '%s'... will try to sync with 'time' command", result)
	//just to make sure the server is ok
	result, err = SendMongrelControl(TIME_CMD, TWO_SECS_IN_MICROS, path, ctx)
	if err != nil {
		return err
	}
	config.Logger.Printf("[ZMQ] time successful: '%s'", result)
	return nil
}

func SendMongrelControl(cmd string, wait int64, path string, ctx gozmq.Context) (string, error) {

	//create ZMQ socket
	s, err := ctx.NewSocket(gozmq.REQ)
	if err != nil {
		return "", err
	}

	//defer it, but be careful in case it got closed early
	defer func() {
		if s != nil {
			s.Close()
		}
	}()

	//immediate death
	err = s.SetSockOptInt(gozmq.LINGER, 0)
	if err != nil {
		return "", err
	}

	//connect to it... but not guaranteed somebody is listening
	err = s.Connect(path)
	if err != nil {
		return "", err
	}

	//we can do the send, whether anyone is listening or not
	err = s.Send([]byte(cmd), 0)
	if err != nil {
		return "", err
	}

	//set for a poll to see if the server responded
	items := make([]gozmq.PollItem, 1)
	items[0].Socket = s
	items[0].Events = gozmq.POLLIN
	//run the poll
	count, err := gozmq.Poll(items, wait)
	if err != nil {
		return "", err
	}
	if count != 0 {
		//config.Logger.Printf("zmq poll set the flags... POLLIN=%v\n", (items[0].REvents&gozmq.POLLIN) != 0)
		if items[0].REvents&gozmq.POLLIN == 0 {
			return "", errors.New("Unable explain why poll returned an item, but it was not POLLIN!")
		}
	} else {
		//we close it now because we don't want mongrel getting the message when it starts up!
		//we us the =nil signal on s to prevent problems with our defered close above
		s.Close()
		s = nil
		return "", nil
	}

	resp, err := s.Recv(0)
	if err != nil {
		return "", err
	}

	return string(resp), nil
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
	`CREATE TABLE IF NOT EXISTS mimetype (id INTEGER PRIMARY KEY, mimetype TEXT, extension TEXT);`,
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
