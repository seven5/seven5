package seven5

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type PagegenOpts struct {
	Funcs           map[string]interface{} //but they better be funcs
	BaseDir         string                 //normally cmd line opt
	SupportDir      string                 //normally cmd line opt
	JsonSupportFile string
	JsonFile        string
	TemplateFile    string
	Debug           bool
	TemplateSuffix  string
}

func (po PagegenOpts) debugf(s string, i ...interface{}) {
	if po.Debug {
		log.Printf("DEBUG:"+s, i...)
	}
}

func (po PagegenOpts) readAll(path string) string {
	f, err := os.Open(filepath.Join(po.BaseDir, path))
	if err != nil {
		log.Fatalf("%s: %v", filepath.Join(po.BaseDir, path), err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("%s (read all): %v", filepath.Join(po.BaseDir, path), err)
	}
	f.Close()
	return string(b)
}

func (po PagegenOpts) newTemplate(s string) *template.Template {
	if len(po.Funcs) > 0 {
		return template.New(s).Funcs(po.Funcs)
	}
	return template.New(s)
}

func (po PagegenOpts) confirmDir(path string) {
	po.confirmDirOrFile(path, true)
}

func (po PagegenOpts) confirmFile(path string) {
	po.confirmDirOrFile(path, false)
}

func (po PagegenOpts) confirmDirOrFile(path string, isDir bool) {
	info, err := os.Stat(path)
	if err != nil {
		log.Fatalf("Unable to find file %s: %v ", path, err)
	}
	if info.IsDir() != isDir {
		if isDir {
			log.Fatalf("%s is file, not a directory ", path)
		} else {
			log.Fatalf("%s is directory, not a file ", path)
		}
	}
}

func (po PagegenOpts) Main() {
	po.confirmDir(po.BaseDir)
	if po.SupportDir != "" {
		po.confirmDir(filepath.Join(po.BaseDir, po.SupportDir))
	}
	po.confirmFile(filepath.Join(po.BaseDir, po.TemplateFile))
	if po.JsonFile != "" {
		po.confirmFile(filepath.Join(po.BaseDir, po.JsonFile))
	}
	if po.JsonSupportFile != "" {
		if filepath.Ext(po.JsonSupportFile) != ".tmpl" {
			log.Fatalf("json support file should be a template (.tmpl extension): %s", po.JsonSupportFile)
		}
		po.confirmFile(filepath.Join(po.BaseDir, po.JsonSupportFile))
	}
	if po.JsonFile != "" {
		if filepath.Ext(po.JsonFile) != ".json" {
			log.Fatalf("json file should be .json extension: %s", po.JsonFile)
		}
		po.confirmFile(filepath.Join(po.BaseDir, po.JsonSupportFile))
	}
	if po.SupportDir != "" {
		po.confirmDir(filepath.Join(po.BaseDir, po.SupportDir))
	}

	//
	// JSON
	//
	jsonData := "{}"
	if po.JsonFile != "" {
		t := po.newTemplate(po.JsonFile)
		if po.JsonSupportFile != "" {
			po.debugf("parsing support file: %s", filepath.Join(po.BaseDir, po.JsonSupportFile))
			t = template.Must(t.ParseFiles(filepath.Join(po.BaseDir, po.JsonSupportFile)))
		}
		po.debugf("parsing json file: %s", filepath.Join(po.BaseDir, po.JsonFile))
		t = template.Must(t.Parse(po.readAll(po.JsonFile)))
		var buffer bytes.Buffer
		t.ExecuteTemplate(&buffer, po.JsonFile, nil)
		jsonData = buffer.String()
	}
	po.debugf("decoding json data (%d bytes)", len(jsonData))
	data := make(map[string]interface{})
	dec := json.NewDecoder(strings.NewReader(jsonData))
	if err := dec.Decode(&data); err != nil {
		po.debugf("---Failing JSON Data---\n%s\n", jsonData)
		log.Fatalf("Unable to decode data  %s (json badly formed?): %v", po.JsonFile, err)
	}

	//
	// TEMPLATE
	//

	t := po.newTemplate(po.TemplateFile)
	if po.SupportDir != "" {
		po.debugf("parsing all .tmpl files in %s", filepath.Join(po.BaseDir, po.SupportDir))
		dir := filepath.Join(po.BaseDir, po.SupportDir)
		template.Must(t.ParseGlob(filepath.Join(dir, "*.tmpl")))
	}
	po.debugf("parsing template file: %s", filepath.Join(po.BaseDir, po.TemplateFile))
	t = template.Must(t.Parse(po.readAll(po.TemplateFile)))
	t.ExecuteTemplate(os.Stdout, po.TemplateFile, data)

}
