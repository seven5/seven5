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
	//PrivateInit is true if a function called PrivateInit is located in the pwd.go
	//file.  At the moment, it doesn't try to check the types of params or return values,
	//just counts them!
	PrivateInit bool
}

const backbone = ".bbone.go"

//CheapASTAnalysis is a function that does some very weak analysis to 
//fill in the ExportedSeven5Objects object that has been passed in. 
//At the moment this analysis is primitive.
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
		if n=="pwd.go" {
			analyzePwd(filepath.Join(dirname,n),exported)
		}
	}

}

//analyzePwd looks for the PrivateInit method in pwd.go files. It checks only that the function
//exits and has the right number of parameters and number of return values (2 and 1).
func analyzePwd(path string, exported *ExportedSeven5Objects) {
	var fset token.FileSet
	var file *ast.File
	var err error
	
	if file, err = parser.ParseFile(&fset, path, nil, parser.ParseComments); err != nil {
		fmt.Fprintf(os.Stderr, "bad parse of %s:%s\n", path, err)
		return
	}
	
	
	for _, exp := range file.Scope.Objects {
		if exp.Kind != ast.Fun{
			continue
		}
		if exp.Name!="PrivateInit" {
			continue
		}		
		fun, ok := exp.Decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		params:=fun.Type.Params;
		results:=fun.Type.Results

		if len(params.List)==2 && len(results.List)==1 {
			exported.PrivateInit=true;
			return
		}
		
		fmt.Fprintf(os.Stderr,"Warning: you have a PrivateInit method but it does not have the correct signature (%d,%d)! Ignored.\n", len(params.List),len(results.List));
	}
	return
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
		ast.Walk(&modelVisitor{exported,exp.Name,fset},decl.Type)
	}
}

//modelVisitor is used to visit each of the nodes of the structure to see if it has any names
//that are going to be a problem in Javascript.
type modelVisitor struct {
	exported *ExportedSeven5Objects
	name string
	fset token.FileSet
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
		var fullName string
		
		switch tnameType:=tname.(type) {
		case *ast.Ident:
			fullName=tnameType.Name
		case *ast.SelectorExpr:
			fullName=tnameType.X.(*ast.Ident).Name+"_"+tnameType.Sel.Name
		default:
			panic(fmt.Sprintf("type of field %v is too complex for us to understand!",f.Names))
		}
		
		for _,candidate := range n {
			if !isNameOkForJS(candidate) {
				fmt.Fprintf(os.Stderr,"Warning: Ignoring %s type because %s is not translatable to Javascript (name)!\n",self.name, candidate)
				return nil 
			}
		}
		if !isTypeOkForJS(fullName) {
			fmt.Fprintf(os.Stderr,"Warning: Ignoring %s type because %s is not translatable to Javascript (typename)!\n",self.name,tname)
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
func isTypeOkForJS(s string) bool {
	return true
}
