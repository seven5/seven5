package seven5

import (
	"errors"
	"exp/sql"
	"flag"
	"fmt"
	"github.com/seven5/gozmq"
	_ "github.com/mattn/go-sqlite3" //force linkage,without naming the library
	"log"
	"github.com/seven5/mongrel2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

//projectConfig represents the configuration of a Seven5 web app.  
type projectConfig struct {
	Path     string
	Name     string
	Logger   *log.Logger
	Addr     []*mongrel2.HandlerSpec
	Service  []string
	ServerId string
	Db       *sql.DB
}

const (
	test_server_id = "f548a1be-3c3c-4f0d-91bd-44edf672af31"

	logdir   = "log"
	mongrel2dir = "mongrel2"
	our_log  = "seven5.log"

	//WEBAPP_START_DIR is _seven5 and must be exported so the "tune" program knows where to put the
	//code it generates.
	WEBAPP_START_DIR = "_seven5"

	localhost  = "localhost"
	access_log = "/" + logdir + "/access.log"
	error_log  = "/" + logdir + "/error.log"
	pid_file   = run_dir + "/mongrel2.pid"
	test_port  = 6767
	run_dir        = "/run"
	control    = "control"

	time_cmd                = "10:4:time,0:}]"
	reload_cmd              = "12:6:reload,0:}]"
	two_secs_in_micros      = 2000000
	hundred_msecs_in_micros = 100000

	static_dir = "static/" //must have trailing slash

	no_such_table = "no such table"

	handler_suffix = "_rawhttp.go"

	handler_insert_sql = `insert into handler(send_spec,send_ident,recv_spec,recv_ident) values("%s","%s","%s","");`
	host_insert_sql               = `insert into host(server_id,name,matching) values(last_insert_rowid(),"%s","%s");`
	server_insert_sql             = `insert into server(uuid,access_log,error_log,pid_file,chroot,default_host,name,port) values("%s","%s","%s","%s","%s","%s","%s","%d");`
	mime_insert_sql               = `insert into mimetype(mimetype,extension) values("%s","%s");`
	directory_insert_sql          = `insert into directory(base,index_file,default_ctype) values("%s","index.html","text/plain");`
	route_insert_sql              = `insert into route(path,host_id,target_id,target_type) values("%s",%s, %s, "%s");`
	log_insert_sql                = `insert into log(who,what,location,how,why) values("user","load", "localhost", "%s","webapp_start");`
)

//verifyProjectLayout checks that the directory structure that is expected for
//a correct Seven5 application is present.  
func verifyProjectLayout(projectPath string) string {
	for _, dir := range []string{logdir, run_dir, static_dir} {
		candidate := filepath.Join(projectPath, mongrel2dir, dir)
		clean := filepath.Clean(candidate)
		if s, _ := os.Stat(clean); s == nil || !s.IsDir() {
			return fmt.Sprintf("Unable to find directory %s (%v)", clean, s == nil)
		}
	}
	return ""
}

//createLogger builds a logger that is connected to the "standard" place in 
//a Seven5 application (mongrel2/log/seven5.log)
func createLogger(projectPath string) (*log.Logger, string, error) {

	path := filepath.Join(projectPath, mongrel2dir, logdir, our_log)
	file, err := os.Create(path)
	if err != nil {
		return nil, "", err
	}

	result := log.New(file, "", log.LstdFlags|log.Lshortfile)
	return result, path, nil
}

//newPorjectConfig creates a projectConfig object and return a pointer to it.  This method
//fills in some fields that are already known such as the path, name, and 
//logger.
func newprojectConfig(path string, l *log.Logger) (*projectConfig, error) {
	result := new(projectConfig)
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

//clearTestDB opens the sqlite3 database for mongrel3 configuration and insures
//that the tables are present and empty.  If the tables are not present, this
//function creates them.
func clearTestDB(config *projectConfig) error {

	db, err := sql.Open("sqlite3", db_path(config))
	if err != nil {
		return err
	}

	config.Logger.Printf("[CONFIG] created/found mongrel2 configuration db at path: %s", db_path(config))

	//destroy all tables, create from scratch
	for _, create := range tabledefs_sql {
		_, err = db.Exec(create)
		if err != nil {
			config.Logger.Printf("[ERROR!] unable to create tables in mongrel2 config:%s", err.Error())
			return err
		}
	}

	config.ServerId = test_server_id

	return nil
}

//generateServerConfig writes information about the server (in mongrel2 terms) name and what host it is
//working on behalf of.  It also configures information about the log file placement
//and other such variables.  The optional arguments are
//accessLog string, errorLog string, pidFile string, chroot string. The "host" portion of the
//mongrel2 configuration is made to be exactly one host whose name is default host.
func generateServerHostConfig(config *projectConfig, defaultHost string, port int, opt ...string) error {

	if len(opt) != 4 && len(opt) != 0 {
		return errors.New("GenerateServerConfig: wrong number of arguments!")
	}
	var chroot, accessLog, errorLog, pidFile string
	if len(opt) == 0 {
		chroot = filepath.Join(config.Path, mongrel2dir)
		accessLog = access_log
		errorLog = error_log
		pidFile = pid_file
	} else {
		accessLog = opt[0]
		errorLog = opt[1]
		pidFile = opt[2]
		chroot = opt[3]
	}
	sqlText := fmt.Sprintf(server_insert_sql, config.ServerId, accessLog, errorLog, pidFile, chroot, defaultHost, config.Name, port)
	r, err := config.Db.Exec(sqlText)
	if err != nil {
		return err
	}
	res, err := r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[mongrel2dir SQL] inserted server into config:%s (%d row)\n", sqlText, res)

	//HOST: must run IMMEDIATELY after the server is inserted because sql is dependent on order
	sqlText = fmt.Sprintf(host_insert_sql, defaultHost, defaultHost)
	r, err = config.Db.Exec(sqlText)
	if err != nil {
		return err
	}
	res, err = r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[mongrel2dir SQL] inserted host into config:%s (%d row)\n", sqlText, res)
	return nil
}

//generateHandlerAddressAndRouteConfig creates the necessary mongrel2 configuration for a handler
//with a given name. It creates an address for the handler, an ID for it and writes those
//into the mongrel2 configuration.  Then it creates a mongrel2 route for the handler to 
//be contacted on.  The route is @[name] if the handler is a JSON service, otherwise it is
//bound into the HTTP space at /[name]
func generateHandlerAddressAndRouteConfig(config *projectConfig, host string, handler Routable) error {
	addr, err := mongrel2.GetHandlerSpec(handler.Name())
	if err != nil {
		return err
	}
	sqlText := fmt.Sprintf(handler_insert_sql, addr.PullSpec, addr.Identity, addr.PubSpec)
	r, err := config.Db.Exec(sqlText)
	if err != nil {
		return err
	}
	res, err := r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[mongrel2dir SQL] inserted %s handler configuration:%s (%d row)\n", addr.Name, sqlText, res)

	//this is the query used to find the host that routes point at
	nestedHost := fmt.Sprintf(`(select id from host where name="%s")`, host)

	// ROUTE TO HANDLER
	nestedHandler := fmt.Sprintf(`(select id from handler where send_spec="%s")`, addr.PullSpec)
	route := "/api/"+handler.Name();
	if (handler.Pattern()!="") {
		route=handler.Pattern();
	}
	sqlText = fmt.Sprintf(route_insert_sql, route, nestedHost, nestedHandler, "handler")
	r, err = config.Db.Exec(sqlText)
	if err != nil {
		return err
	}
	res, err = r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[mongrel2dir SQL] inserted handler %s into routes:%s (%d rows affected)\n", addr.Name, sqlText, res)
	return nil
}

//generateStaticContentConfigcreates the necessary mongrel2 configuration for a directory
//a 'directory' and a 'route' to that directory.
func generateStaticContentConfig(config *projectConfig, host string, path string) error {
	//this is the query used to find the host that routes point at
	nestedHost := fmt.Sprintf(`(select id from host where name="%s")`, host)

	dirText := fmt.Sprintf(directory_insert_sql, path)
	r, err := config.Db.Exec(dirText)
	if err != nil {
		return err
	}
	res, err := r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[mongrel2dir SQL] inserted directory into config:%s (%d row)\n", dirText, res)

	// ROUTE TO static_dir CONTENT
	staticDirectory := fmt.Sprintf(`(select id from directory where base="%s")`, static_dir)
	routeText := fmt.Sprintf(route_insert_sql, "/static/", nestedHost, staticDirectory, "dir")

	r, err = config.Db.Exec(routeText)
	if err != nil {
		return err
	}
	res, err = r.RowsAffected()
	if err != nil {
		return err
	}
	config.Logger.Printf("[mongrel2dir SQL] inserted static content route into config:%s (%d row)\n", routeText, res)
	return nil
}

//generateMimeTypeConfig puts the mime-type table in place if it is needed.
func generateMimeTypeConfig(config *projectConfig) error {
	/*
	rows, err := config.Db.Query("select count(*) from mimetype;")
	if err != nil {
		return err
	}

	var result int
	rows.Next()
	rows.Scan(&result)
	//XXX mattn needs to fix something in the sqlite3 support for Close() to work XXX
	rows.Close()
	*/
	
	/*temporary workaround just create the table every time*/
	
	result:=0
	_, err := config.Db.Exec("DROP TABLE IF EXISTS mimetype;")
	if err != nil {
		return err
	}
	_, err = config.Db.Exec("CREATE TABLE IF NOT EXISTS mimetype (id INTEGER PRIMARY KEY, mimetype TEXT, extension TEXT);")
	if err != nil {
		return err
	}

	/*end of temporary workaround*/

	config.Logger.Printf("[mongrel2dir SQL] currenty %d items in mimetype table\n", result)

	if result == 0 {
		for _, pair := range mime_type {
			mimeText := fmt.Sprintf(mime_insert_sql, pair[0], pair[1])
			_, err = config.Db.Exec(mimeText)
			if err != nil {
				return err
			}
		}
		config.Logger.Printf("[mongrel2dir SQL] inserted %d mime types in mongrel2 configuration\n", len(mime_type))
	}
	return nil
}

//bootstrap should be called by projects that use the Seven5 infrastructure to
//parse the command line arguments and to discover the project's structure.
//This function returns null if the project config is not standard.  This is a 
//wrapper on bootstrapFromDir that gets the project directory from the command line args. 
func bootstrap() *projectConfig {

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
	return bootstrapFromDir(projectDir)
}

//bootstrapFromDir does the heavy lifting to set up a project, given a directory to work
//with.  It returns a configuration (or nil on error).
func bootstrapFromDir(projectDir string) *projectConfig {
	if err := verifyProjectLayout(projectDir); err != "" {
		dumpBadProjectLayout(projectDir, err)
		return nil
	}

	logger, path, err := createLogger(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create logger:%s\n", err)
		return nil
	}

	fmt.Printf("Seven5 is logging to %s\n", path)

	config, err := newprojectConfig(projectDir, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to compute project configuration (problem with path to configuration db?):%s\n", err.Error())
		return nil
	}

	config.Logger.Printf("---- PROJECT %s @ %s ----", config.Name, config.Path)

	db, err := sql.Open("sqlite3", db_path(config))
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to open configuration database:%s\n", err.Error())
		return nil
	}
	config.Db = db

	err = clearTestDB(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error clearing the test db configuration:%s\n", err.Error())
		return nil
	}

	return config
}

//createNetworkResources creates the necessary 0MQ context and starts or re-initializes
//the mongrel2 server.  It returns the context if everything is ok and this should be
//used by the whole program.
func createNetworkResources(config *projectConfig) (Transport, error) {
	ctx, err := gozmq.NewContext()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create zmq context :%s\n", err.Error())
		return nil, nil
	}

	err = runMongrel(config, ctx)
	//fmt.Fprintf(os.Stderr, "Ctx %v err %v\n", ctx, err)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to start/reset mongrel2:%s\n", err.Error())
		return nil, nil
	}

	return NewTransport(ctx), nil

}

//startMongrel is a support routine that knows how to start a mongrel2 instance from the path and
//set the error log and output log to the standard seven5 places.
func startMongrel(config *projectConfig, ctx gozmq.Context) error {

	mongrelPath, err := exec.LookPath("mongrel2")
	if err != nil {
		config.Logger.Printf("[ERROR!] unable to find mongrel2 in path! you should probably change your PATH environment variable")
		return err
	}

	config.Logger.Printf("[PROCESS] using mongrel2 binary at %s", mongrelPath)

	stdout, err := os.Create(filepath.Join(config.Path, mongrel2dir, logdir, "mongrel2.out.log"))
	if err != nil {
		return err
	}

	stderr, err := os.Create(filepath.Join(config.Path, mongrel2dir, logdir, "mongrel2.err.log"))
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

	socketPath := filepath.Join(config.Path, mongrel2dir, run_dir, control)
	path := fmt.Sprintf("ipc://%s", socketPath)
	config.Logger.Printf("[ZMQ] using zmq connection to mongrel2 for sync with startup '%s'", path)
	response, err := sendMongrelControl(time_cmd, two_secs_in_micros, path, config, ctx)
	if err != nil {
		return err
	}
	config.Logger.Printf("[ZMQ] time command on startup succeeded, response was '%s'", response)

	return nil
}

//runMongrel is a support routine that can detect if a mongrel2 is running in the standard place
//and verify that it is running correctly.  If it succeeds at doing this job, there is nothing
//else to do.  If it fails, it calls startMongrel to try to bring up a new copy of mongrel2.
func runMongrel(config *projectConfig, ctx gozmq.Context) error {

	socketPath := filepath.Join(config.Path, mongrel2dir, run_dir, control)
	path := fmt.Sprintf("ipc://%s", socketPath)
	config.Logger.Printf("[ZMQ] zmq connection to mongrel2 is '%s'", path)

	result, err := sendMongrelControl(reload_cmd, hundred_msecs_in_micros, path, config, ctx)
	if err != nil {
		return err
	}
	if result == "" {
		config.Logger.Printf("[ZMQ] did not get any response on mongrel control connection... will try to start mongrel")
		return startMongrel(config, ctx)
	}
	config.Logger.Printf("[ZMQ] reload successful: '%s'... will try to sync with 'time' command", result)
	//just to make sure the server is ok
	result, err = sendMongrelControl(time_cmd, two_secs_in_micros, path, config, ctx)
	if err != nil {
		return err
	}
	if strings.Trim(result," ")=="" {
		config.Logger.Printf("[ZMQ] WARNING! time command seemed to succeed, but empty result!")
	} else {
		config.Logger.Printf("[ZMQ] time successful: '%s'", result)
	}
	return nil
}

//sendMongrelControl is the routine for talking to the control port of mongrel2.  It knows how to send
//a command and get the response (these use netstrings).
func sendMongrelControl(cmd string, wait int64, path string, config *projectConfig, ctx gozmq.Context) (string, error) {

	//create ZMQ socket
	s, err := ctx.NewSocket(gozmq.REQ)
	if err != nil {
    	config.Logger.Printf("[ERROR!] could not create a new socket: '%v'", err)
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
    	config.Logger.Printf("[ERROR!] could not set socket int %v: '%v'", gozmq.LINGER, err)
		return "", err
	}

	//connect to it... but not guaranteed somebody is listening
	err = s.Connect(path)
	if err != nil {
    	config.Logger.Printf("[ERROR!] could not connect to path: '%v'", path)
		return "", err
	}

	//we can do the send, whether anyone is listening or not
	err = s.Send([]byte(cmd), 0)
	if err != nil {
    	config.Logger.Printf("[ERROR!] could not send the command: '%v'", cmd)
		return "", err
	}

	//set for a poll to see if the server responded
	items := make([]gozmq.PollItem, 1)
	items[0].Socket = s
	items[0].Events = gozmq.POLLIN
	//run the poll
	count, err := gozmq.Poll(items, wait)
	if err != nil {
    	config.Logger.Print("[ERROR!] could not POLL")
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
    	config.Logger.Printf("[ERROR!] could not receive from the socket: '%v'", err)
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
	fmt.Fprintf(os.Stderr, "\nfor project structure details, see http://seven5.github.com/seven5/develop.html\n\n")
}

//db_path is a support routine to return the path to the SQLite db, given a project structure
func db_path(config *projectConfig) string {
	return filepath.Join(config.Path, fmt.Sprintf("%s_test.sqlite", config.Name))

}

//finishConfig completes the process of writing the data into the mongrel2 configuration
//database and releases any resources used.
func finishConfig(config *projectConfig) error {
	return config.Db.Close()
}

//tabledefs_sql is the SQL to create the mongrel2 tables.
var tabledefs_sql = []string{
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
