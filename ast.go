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

//ExportedSeven5Objects is the information that we have "discovered" by analyzing the AST of the
//a user project. This is passed to the "tune" program so it is public.
type ExportedSeven5Objects struct {
	//Model is the names of the exported structs
	Model []string
}

const backbone = ".bbone.go"

//CheapASTAnalysis is a function that does some very weak analysis to try to find models that are
//being exported by a user-level package.  If it finds them, it puts them in the exported parameter.
//This is called by the "tune" program so it is public.
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
		if strings.HasSuffix(n,backbone) {
			analyzeBackbone(filepath.Join(dirname,n),exported)
			continue
		}
	}

}

//analyzeBackbone looks for declarations that smell like a backbone model.  if it finds them,
//it addes them to the exported object.  Currently it just looks for structs with exported names.
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

//modelVisitor is used to visit each of the nodes of the structure to see if it has any names
//that are going to be a problem in Javascript.
type modelVisitor struct {
	exported *ExportedSeven5Objects
	name string
}

//Visit does the work for modelVisitor of examining all the fields of a struct looking for names that
//are not ok for Javascript.  If it finds a name that is not ok in javascript it prints an error
//and the wanna-be model is ignored entirely.
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
		self.exported.Model=[]string{}
	} 
	self.exported.Model=append(self.exported.Model,self.name)
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
