package seven5

import (
	"encoding/json"
	"seven5/util"
	"strings"
)

const (
	GROUPIE_CONFIG_FILE = "groupie.json"
)

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

type GroupieWrapper struct {
	GroupieConfig []Groupie
}

//GroupieConfig is the result of parsing the json.
type groupieConfig map[string]*GroupieInfo

//parseGroupieConfig takes a bunch of json and turns it into a groupieConfig.
//It returns an error if you don't supply a sensible configuration.
func parseGroupieConfig(jsonBlob string) (groupieConfig, error) {
	result := make(map[string]*GroupieInfo)
	dec := json.NewDecoder(strings.NewReader(jsonBlob))
	var wrapper GroupieWrapper
	if err := dec.Decode(&wrapper); err != nil {
		return nil, err
	} 
	for _, raw := range wrapper.GroupieConfig {
		result[raw.Role] = &GroupieInfo{raw.Info.TypeName, raw.Info.ImportsNeeded}
	}
	return result, nil
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
