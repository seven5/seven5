package seven5

import (
	"os"
	"net/http"
	"encoding/json"
	"strings"
	"path/filepath"
	"fmt"
)
//publicsetting is used to hold the mapping from keys to values used for secrets to be exposed by the server
//if requested by the client.
var publicsetting map[string]string

//publicSettingReader reads the filename provided in the project go directory to get the list of public settings.
//The file's content should be a json map from strings to strings.  This is read only once at startup
//time.
func publicSettingReader(projectName string, filename string) error {
	path, err:=ProjRootDirFromPath(projectName)
	if err!=nil {
		return err
	}
	f, err:= os.Open(filepath.Join(path, filename))
	if err!=nil {
		return err
	}
	defer f.Close()
	dec:=json.NewDecoder(f)
	if err=dec.Decode(&publicsetting); err!=nil {
		return err
	}
	return nil
}

//publicSettingHandler responds to requests for a publicsetting or gives 404. 
//public settings are placed in publicsetting.json
func publicSettingHandler(response http.ResponseWriter, request *http.Request) {
	found := false
	parts := strings.Split(request.URL.Path, "/")
	i := 0
	targ := 0
	projectName :=0
	for ; i < len(parts); i++ {
		if parts[i] == "publicsetting" {
			targ = i + 2
			projectName = i + 1
			found = true
			break
		}
	}
	if !found || (found && targ >= len(parts)) {
		http.Error(response, "can't understand path", http.StatusNotFound)
		return
	}
	if parts[targ] == "" {
		http.Error(response, "no public setting requested", http.StatusBadRequest)
		return
	}
	if parts[projectName] == "" {
		http.Error(response, "no project name requested", http.StatusBadRequest)
		return
	}
	if publicsetting==nil {
		publicsetting=make(map[string]string)
		if err:=publicSettingReader(parts[projectName], "publicsetting.json"); err!=nil {
			publicsetting=nil //try again if they fix the formatting
			http.Error(response, fmt.Sprintf("failed reading public settings file: %s", err), 
				http.StatusInternalServerError)
			return
		}
	}	
	
	s, ok:= publicsetting[parts[targ]]
	if !ok {
		http.Error(response, "public setting not found", http.StatusNotFound)
	}
	response.Write([]byte(s))
}
