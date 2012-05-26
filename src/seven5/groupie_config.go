package seven5

import (
	"seven5/util"
	"encoding/json"
	"strings"
	"io"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"bytes"
)

const (
	GROUPIE_CONFIG_FILE="groupie.json"
)

//substitute for constant array
func MANDATORY_ROLES() []string {
	return []string{ "ProjectValidator"}
}

//substitute for constant array
func ALL_ROLES() []string {
	return []string{ "ProjectValidator"}
}

// Groupie plays a role in the system. These roles are well defined and
// bound to structures that will be passed to the groupie playing that
// role.
type Groupie struct {
	Role string
	Info GroupieInfo
}

//GroupieInfo is extra information about a particular groupie. This is 
//information used/needed at bootstrap time.
type GroupieInfo struct {
	TypeName string
	ImportsNeeded []string
}

//GroupieConfig is the result of parsing the json
type GroupieConfig map[string]*GroupieInfo


//ParseGroupieConfig takes a bunch of JSON and turns it into a GroupieConfig
func ParseGroupieConfig(jsonBlob string) (GroupieConfig, error) {
	result := make(map[string]*GroupieInfo)
	mandatory := util.NewBetterList()
	for _,k := range(MANDATORY_ROLES()) {
		mandatory.PushBack(k)
	}
	possible := util.NewBetterList()
	for _,k := range(ALL_ROLES()) {
		possible.PushBack(k)
	}
	dec := json.NewDecoder(strings.NewReader(jsonBlob))
	for {
		var groupie Groupie
	    if err := dec.Decode(&groupie); err == io.EOF {
    	    break
    	} else if err != nil {
        	return nil,errors.New(fmt.Sprintf("Cannot understand JSON in %s",
        		GROUPIE_CONFIG_FILE))
        }
        //basic sanity check
        if !possible.Contains(groupie.Role) {
        	return nil,errors.New(fmt.Sprintf("Reading %s and found unknown groupie role: %s",
        		GROUPIE_CONFIG_FILE, groupie.Role))
        }
        if !possible.Contains(groupie.Role) {
        	return nil,errors.New(fmt.Sprintf("Reading %s and found groupie role multiple times: %s",
        		GROUPIE_CONFIG_FILE, groupie.Role))
        }
        result[groupie.Role]=&groupie.Info
        possible.RemoveValue(groupie.Role)
    }
    if possible.Len()!=0 {
        	return nil, errors.New(fmt.Sprintf("Reading %s did not find all mandatory roles: %s",
        		GROUPIE_CONFIG_FILE,possible.AllValues()))
    }
    return result,nil
}

//FindGroupieConfigFile returns a String with the contents of the config file
//in the current directory.
func FindGroupieConfigFile() (string, error){
	var cwd string
	var err error
	var file *os.File
	
	//get cwd typically of groupie
	if cwd, err = os.Getwd(); err != nil {
		return "",err
	}

	configPath := filepath.Join(cwd, "groupies.json")

	if file, err = os.Open(configPath); err != nil {
		return "",err
	}

	var jsonBuffer bytes.Buffer
	if _, err = jsonBuffer.ReadFrom(file); err != nil {
		return "", err
	}
	
	return jsonBuffer.String(),nil

}
