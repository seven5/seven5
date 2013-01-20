package seven5tool

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var Bigidea = &Command{bigideaFn, "bigidea",
	"create a new project in a given directory",
	`
bigidea is used to create the default seven5 project structure in a given directory.
It is careful not to blow away the directory if it already exists.  It also emits to
the terminal an example setting for the GOPATH command.
`,
}

func bigideaFn(argv []string, l *log.Logger) {

	fset := flag.NewFlagSet("bigideaFlags", flag.ExitOnError)
	fset.Parse(argv)
	if len(fset.Args()) != 1 {
		l.Printf("must supply exactly one directory (not %d), the directory where the new project is to be created", len(fset.Args()))
		return
	}

	idea := filepath.Clean(fset.Arg(0))

	cwd, err := os.Getwd()
	if err != nil {
		l.Printf("failed to get working directory:%s", err)
		return
	}

	if !filepath.IsAbs(idea) {
		idea = filepath.Join(cwd, idea)
	}

	cand, err := os.Open(idea)
	if err == nil {
		l.Printf("Not overwriting %s!", idea)
		return
	}

	if cand != nil {
		l.Printf("Whoa! Non-nil return value from bad os.Open(%s)!", idea)
		return
	}

	if err = os.Mkdir(idea, os.ModeDir|os.ModePerm); err != nil {
		l.Printf("Can't create directory %s! %s", idea, err)
		return
	}

	for _, d := range []string{"dart", "go", "static"} {
		p := filepath.Join(idea, d)
		if err = os.Mkdir(p, os.ModeDir|os.ModePerm); err != nil {
			l.Printf("Can't create directory %s! %s", p, err)
			return
		}
	}

	for _, d := range []string{"src", "bin", "pkg"} {
		p := filepath.Join(idea, "go", d)
		if err = os.Mkdir(p, os.ModeDir|os.ModePerm); err != nil {
			l.Printf("Can't create directory %s! %s", p, err)
			return
		}
	}

	base := filepath.Base(idea)
	for _, c := range base {
		if c == ' ' || c == '.' {
			l.Printf("base name for project is not ok: cannot contain spaces or dots: %s", base)
			return
		}
	}

	//write the sample dart code
	sample, err := os.Create(filepath.Join(idea, "dart", base+".dart"))
	if err != nil {
		l.Printf("Can't create sample dart ui code %s! %s", filepath.Join(idea, "dart", base+".dart"), err)
		return
	}

	_, err = sample.Write([]byte(ui_dart))
	if err != nil {
		l.Printf("Can't write sample dart ui code %s! %s", filepath.Join(idea, "dart", base+".dart"), err)
		return
	}

	if err = os.Mkdir(filepath.Join(idea, "go", "src", base), os.ModeDir|os.ModePerm); err != nil {
		l.Printf("Can't create directory %s! %s", filepath.Join(idea, "go", "src", base), err)
		return
	}

	if err = os.Mkdir(filepath.Join(idea, "go", "src", base, "run"+base), os.ModeDir|os.ModePerm); err != nil {
		l.Printf("Can't create directory %s! %s", filepath.Join(idea, "go", "src", base, "run"+base), err)
		return
	}

	lib := template.Must(template.New("lib").Parse(lib_tmpl))
	test := template.Must(template.New("test").Parse(lib_test_tmpl))
	main := template.Must(template.New("main").Parse(main_tmpl))

	templates := []*template.Template{lib, test, main}
	names := []string{"lib", "test", "main"}
	paths := []string{
		filepath.Join(idea, "go", "src", base, base+".go"),
		filepath.Join(idea, "go", "src", base, base+"_test.go"),
		filepath.Join(idea, "go", "src", base, "run"+base, base+"main.go"),
	}

	data := make(map[string]string)
	data["name"] = strings.ToUpper(base[0:1]) + base[1:]
	data["base"] = base

	for i, t := range templates {
		out, err := os.Create(paths[i])
		if err != nil {
			l.Printf("Can't create output file %s:%s", paths[i], err)
			return
		}
		if err := t.ExecuteTemplate(out, names[i], data); err != nil {

			return
		}
	}
	
	l.Printf("project '%s' created.  You probably want to set GOPATH and PATH like this:\n", base)

	gp:=os.Getenv("GOPATH")
	if gp!="" {
		gp=":$GOPATH"
	}
	l.Printf("export GOPATH=%s%s\n", filepath.Join(idea, "go"),gp)
	l.Printf("export PATH=$PATH:%s\n", filepath.Join(idea, "go", "bin"))
}
