package seven5

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"seven5/groupie"
	"seven5/util"
	"strings"
)

const (
	GROUPIE_CONFIG_FILE = "groupie.json"
)

//substitute for constant array
func MANDATORY_ROLES() []string {
	return []string{
		groupie.VALIDATEPROJECT,
		groupie.PROCESSCONTROLLER,
	}
}

//substitute for constant array
func ALL_ROLES() []string {
	return []string{
		groupie.VALIDATEPROJECT,
		groupie.ECHO,
		groupie.PROCESSCONTROLLER,
	}
}

// Groupie plays a role in the system. These roles are well defined and
// bound to structures that will be passed to the groupie playing that
// role.  Public for json-encoding.
type Groupie struct {
	Role string
	Info GroupieInfo
}

//GroupieInfo is extra information about a particular groupie. This is 
//information used/needed at bootstrap time.  Public for json encoding.
type GroupieInfo struct {
	TypeName      string
	ImportsNeeded []string
}

//GroupieConfig is the result of parsing the json.
type groupieConfig map[string]*GroupieInfo

//parseGroupieConfig takes a bunch of json and turns it into a groupieConfig.
//It returns an error if you don't supply a sensible configuration.
func parseGroupieConfig(jsonBlob string) (groupieConfig, error) {
	result := make(map[string]*GroupieInfo)
	mandatory := util.NewBetterList()
	for _, k := range MANDATORY_ROLES() {
		mandatory.PushBack(k)
	}
	possible := util.NewBetterList()
	for _, k := range ALL_ROLES() {
		possible.PushBack(k)
	}
	dec := json.NewDecoder(strings.NewReader(jsonBlob))
	for {
		var groupie Groupie
		if err := dec.Decode(&groupie); err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.New(fmt.Sprintf("Cannot understand JSON in %s",
				GROUPIE_CONFIG_FILE))
		}
		//basic sanity check
		if !possible.Contains(groupie.Role) {
			return nil, errors.New(fmt.Sprintf("Reading %s and found unknown groupie role: %s",
				GROUPIE_CONFIG_FILE, groupie.Role))
		}
		if !possible.Contains(groupie.Role) {
			return nil, errors.New(fmt.Sprintf("Reading %s and found groupie role multiple times: %s",
				GROUPIE_CONFIG_FILE, groupie.Role))
		}
		result[groupie.Role] = &groupie.Info
		possible.RemoveValue(groupie.Role)
		mandatory.RemoveValue(groupie.Role)
	}
	if mandatory.Len() != 0 {
		return nil, errors.New(fmt.Sprintf("Reading %s did not find all mandatory roles: %s",
			GROUPIE_CONFIG_FILE, possible.AllValues()))
	}
	return result, nil
}

//findGroupieConfigFile returns a String with the contents of the config file
//in the current directory.  Pass "" to have use current working dir.
func findGroupieConfigFile(cwd string) (string, error) {
	var err error
	var file *os.File

	//get cwd typically of groupie
	if cwd == "" {
		if cwd, err = os.Getwd(); err != nil {
			return "", err
		}
	}

	configPath := filepath.Join(cwd, "groupies.json")

	if file, err = os.Open(configPath); err != nil {
		return "", err
	}

	var jsonBuffer bytes.Buffer
	if _, err = jsonBuffer.ReadFrom(file); err != nil {
		return "", err
	}

	return jsonBuffer.String(), nil

}

// getGroupies is called to read a set of groupie values
// from json to a config structures. It returns nil if the format is not 
// satisfactory (plus an error value).  Note that this does not check semantics!
func getGroupies(jsonBlob string, logger util.SimpleLogger) (groupieConfig, error) {
	var result groupieConfig
	var err error
	if result, err = parseGroupieConfig(jsonBlob); err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	return result, nil
}
