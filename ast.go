package seven5
import (
	//"exp/types"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)


type ExportedSeven5Objects struct {
	Model map[string][]string
}

const BACKBONE = ".bbone.go"

func CheapASTAnalysis(dirname string, exported *ExportedSeven5Objects) {
	var err error
	var dir *os.File
	var names []string
	
	if dir, err = os.Open(dirname); err != nil {
		fmt.Fprintf(os.Stderr, "can't open dir %s:%s\n", dirname, err)
		return
	}

	if names, err = dir.Readdirnames(-1 /*no limit*/ ); err != nil {
		fmt.Fprintf(os.Stderr, "can't read dir names:%s\n", err)
		return
	}

	for _, n := range names {
		if strings.HasSuffix(n,BACKBONE) {
			analyzeBackbone(filepath.Join(dirname,n),exported)
			continue
		}
	}

}

//analyzeBackbone looks for declarations that smell like a backbone model.  if it finds them,
//it addes them to the exported object
func analyzeBackbone(path string, exported *ExportedSeven5Objects) {
	var fset token.FileSet
	var file *ast.File
	var err error
	
	if file, err = parser.ParseFile(&fset, path, nil, parser.ParseComments); err != nil {
		fmt.Fprintf(os.Stderr, "bad parse of %s:%s\n", path, err)
		return
	}
	
	for _, exp := range file.Scope.Objects {
		if exp.Kind != ast.Typ {
			continue
		}
		if !ast.IsExported(exp.Name) {
			continue
		}
		decl, ok := exp.Decl.(*ast.TypeSpec)
		if !ok {
			continue
		}
		_,ok = decl.Type.(*ast.StructType) 
		if !ok {
			continue
		}
		ast.Walk(&modelVisitor{exported,exp.Name},decl.Type)
	}
}

type modelVisitor struct {
	exported *ExportedSeven5Objects
	name string
}

func (self *modelVisitor) Visit(node ast.Node) ast.Visitor {
	t := node.(*ast.StructType)
	result := []string{}
	
	for _,f := range t.Fields.List {
		n:=f.Names
		tname:=f.Type
		
		for _,candidate := range n {
			if !isNameOkForJS(candidate) {
				fmt.Fprintf(os.Stderr,"Warning: Ignoring %s type because %s is not translatable to Javascript!\n",self.name, candidate)
				return nil 
			}
		}
		if !isTypeOkForJS(tname) {
			fmt.Fprintf(os.Stderr,"Warning: Ignoring %s type because %s is not translatable to Javascript!\n",self.name,tname)
			return nil 
		}
		for _,candidate:=range n {
			result=append(result,candidate.Name)
		}
	}
	if self.exported.Model==nil {
		self.exported.Model=make(map[string][]string)
	}
	self.exported.Model[self.name]=result
	return nil
}

//isNameOkForJS for now just checks that we have an identifier and assumes that all go
//identifiers are ok in JS.
func isNameOkForJS(candidate *ast.Ident) bool {
	return true
}

//isTypeOkForJS for now just checks that we have an identifier for the type name and assumes
//that all "direct" types (no pointers or anything) are ok for JS
func isTypeOkForJS(e ast.Expr) bool {
	_, ok:=e.(*ast.Ident)
	return ok
}
