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
	"seven5/dsl"
)


type ExportedSeven5Objects struct {
	CSSId      []string
	CSSClass   []string
	StyleSheet []string
	Document   []string
	Handler    []string
	Model map[string][]string
}

const RAWHTTP = "_rawhttp.go"
const JSONSERVICE = "_jsonservice.go"
const DOTGO = ".go"
const CSS = ".css.go"
const HTML = ".html.go"
const BACKBONE = ".bbone.go"

func CheapASTAnalysis(dirname string, exported *ExportedSeven5Objects) {
	var fset token.FileSet
	//var pkg *ast.Package
	var file *ast.File
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

	files := make(map[string]*ast.File)

	for _, n := range names {
		//check on potential handlers
		if strings.HasSuffix(n, RAWHTTP) || strings.HasSuffix(n, JSONSERVICE) {
			if strings.HasSuffix(n,RAWHTTP) {
				exported.Handler=append(exported.Handler,n[0:len(n)-len(RAWHTTP)])
			} else {
				exported.Handler=append(exported.Handler,n[0:len(n)-len(JSONSERVICE)])
			}
		}
		if strings.HasSuffix(n,BACKBONE) {
			analyzeBackbone(filepath.Join(dirname,n),exported)
			continue
		}
		//from here on, we only do analysis of the DSLs
		if !strings.HasSuffix(n, HTML) && !strings.HasSuffix(n, CSS) {
			continue
		}
		path := filepath.Join(dirname, n)
		if file, err = parser.ParseFile(&fset, path, nil, parser.ParseComments); err != nil {
			fmt.Fprintf(os.Stderr, "bad parse of %s:%s\n", n, err)
			return
		}
		files[n] = file
	}

	for _, f := range files {
		for _, exp := range f.Scope.Objects {
			if exp.Kind != ast.Var {
				continue
			}
			valueSpec, ok := exp.Decl.(*ast.ValueSpec)
			if !ok {
				continue
			}
			ast.Walk(&nameVisitor{exported}, valueSpec)
		}
	}
}

type nameVisitor struct {
	exported *ExportedSeven5Objects
}

func (self *nameVisitor) Visit(node ast.Node) ast.Visitor {
	valueSpec := node.(*ast.ValueSpec)
	names := valueSpec.Names
	if len(valueSpec.Values) == 0 {
		return nil
	}
	for i, v := range valueSpec.Values {
		_, ok := v.(*ast.CallExpr)
		if !ok {
			lit, ok := v.(*ast.CompositeLit)
			if !ok {
				continue
			}
			if !ast.IsExported(names[i].Name) {
				continue
			}
			l := &literalVisitor{names[i].Name,self.exported}
			ast.Walk(l,lit)
		} else {
			if !ast.IsExported(names[i].Name) {
				continue
			}
			c := &callVisitor{names[i].Name,self.exported}
			ast.Walk(c, v)
		}
	}
	return nil
}

type callVisitor struct {
	Name string
	exported *ExportedSeven5Objects
}

func (self *callVisitor) Visit(node ast.Node) ast.Visitor {
	callExpr := node.(*ast.CallExpr)
	f := callExpr.Fun
	i, ok := f.(*ast.Ident)
	if !ok {
		return nil
	}
	switch (i.Name) {
	case "Class":
		self.exported.CSSClass = append(self.exported.CSSClass, self.Name)
	case "Id":
		self.exported.CSSId = append(self.exported.CSSId, self.Name)
	case "Inherit":
		self.exported.CSSClass = append(self.exported.CSSClass, self.Name)
	default:
		fmt.Fprintf(os.Stderr,"seven5 analyzer: ignoring unknown func %s\n",i.Name)
		
	}
	return nil
}

type literalVisitor struct {
	Name string
	exported *ExportedSeven5Objects
}

func (self *literalVisitor) Visit(node ast.Node) ast.Visitor {
	compLit := node.(*ast.CompositeLit)
	i, ok := compLit.Type.(*ast.Ident)
	if !ok {
		return nil
	}
	switch (i.Name) {
	case "StyleSheet":
		self.exported.StyleSheet = append(self.exported.StyleSheet, self.Name)
	case "Document":
		self.exported.Document = append(self.exported.Document, self.Name)
	default:
		fmt.Fprintf(os.Stderr,"seven5 analyzer: ignoring unknown composite literal type %s\n",i.Name)
	}
	return nil
}

func RegisterCSSId(varName string, varValue dsl.Id) {
	
}

func RegisterCSSClass(varName string, varValue dsl.Class) {
	
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