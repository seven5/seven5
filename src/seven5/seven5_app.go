package seven5

//all commands
const (
	VALIDATEPROJECT   = "ValidateProject"
	ECHO              = "Echo"
	PROCESSCONTROLLER = "ProcessController"
	PROCESSVOCAB      = "ProcessVocab"
	BUILDUSERLIB      = "BuildUserLib"
	EXPLODETYPE      = "ExplodeType"
)

//seven5app this is the "application" that is seven5.
var Seven5app = make(map[string]Command)
