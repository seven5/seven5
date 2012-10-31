package seven5tool

import (
	"io"
	"io/ioutil"
	"strings"
	"path/filepath"
	"fmt"
	"log"
	"os"
	"flag"
)

var Embedfile=	&Command{embedfileFn, "embedfile", 
"embed a the text of a file in a go source file as source code",
`
embedfile is used during development when it is necessary to have access to 
particular text files without knowing the path to them at run time.  The 
input file is usually compiled into the binary of an executable go program.
`,
}

func embedfileFn(argv []string, l *log.Logger) {
	var outfile string
	var out io.Writer
	var pkgName string
	var binary bool
	
	fset:=flag.NewFlagSet("embedfileFlags",flag.ExitOnError)
	fset.StringVar(&outfile,"out", "stdout", 
		"path to out file to place source code in [defaults to stdout]")
	fset.StringVar(&pkgName,"pkg", "", 
		"package name to put at the top of the file")
	fset.BoolVar(&binary, "binary",  false, 
		"if you set this flag, the result will be a slice of bytes, not a string")
	fset.Parse(argv)
	if pkgName=="" {
		l.Printf("you must supply a package name to include in the generated file (--pkg=foo)")
	}
	if len(fset.Args())==0 {
		l.Printf("must supply one or more input files to process as text")
		return
	}
	if outfile=="stdout" {
		out=os.Stdout
	} else {
		f, err:=os.Create(outfile)
		if err!=nil {
			l.Printf("%s:%s", outfile, err)
			return
		}
		defer f.Close()
		out=f
	}
	for _, i:=range fset.Args() {
		f, err:=os.Open(i)
		if err!=nil {
			log.Printf("%s:%s",i,err)
			log.Printf("nothing done")
			return
		}
		f.Close()
	}
	//write the package name
	p:=fmt.Sprintf("package %s\n",pkgName)
	out.Write([]byte(p))
	for _, i:=range fset.Args() {
		f, err:=os.Open(i)
		if err!=nil {
			log.Printf("failed to open %s:%s",i,err)
			return
		}
		defer f.Close()
		data, err:=ioutil.ReadAll(f)
		if err!=nil {
			log.Printf("failed to read %s:%s",i,err)
			return
		}
		text:=string(data)
		name:=filepath.Base(i)
		name=strings.Replace(name, ".", "_", -1)
		name=strings.Replace(name, " ", "_", -1)
		if binary {
			decl:=fmt.Sprintf("var %s = []byte{",name)
			out.Write([]byte(decl))
			for i, x:=range data {
				if i%12==0 {
					out.Write([]byte("\n"))
				}
				hex:=fmt.Sprintf("0x%x, ",x)
				out.Write([]byte(hex))
			}
			end:=fmt.Sprintf("\n}\n")
			out.Write([]byte(end))
		} else {
			gocode:=fmt.Sprintf("const %s=`\n%s`\n",name,text)
			out.Write([]byte(gocode))
		}
	}
}



