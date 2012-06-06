package cmd

import (
	"seven5/util"
)

//CommandDecl is the structure for a declaration of a command.  Use a literal
//with named fields for clarity.  Public so others can write commands.
type CommandDecl struct {
	Arg []*CommandArgPair
	Ret *CommandReturn
	Impl func(logger util.SimpleLogger,v...interface{}) interface{}
}

//SimpleReturn is the most common return type for a command, just a boolean
//that indicates if the command ran correctly or not.  This contains all
//the necessary parts of a CommandReturn structure to be used in a command
//declaration. Public so others can write commands.
var SimpleReturn = &CommandReturn {
	Unmarshalled: func() interface{} {return &SimpleErrorReturn{}},
	ErrorTest: simpleErrorTest,
	GetBody: nil,
}
//SimpleBodyReturn is used for commands who want to display output into the 
//web browser (html).  This contains all the necessary parts for a CommandReturn
//to be used in a command declaration.  Public so others can write commands.
var SimpleBodyReturn = &CommandReturn {
	Unmarshalled: func() interface{} {return &BodyReturn{}},
	ErrorTest: bodyErrorTest,
	GetBody: bodyGetBody,
}

//CommandReturn is the structure of a return value from a command. The 
//Unmarshalled function returns an instance of the correct type for use
//in unmarshallin a string argument.  This is used on the client side
//to understand the result returned from the command.  The error test is
//used by the client side to simplify error handling since it can test
//to see if the command returned a semantic error and no further processing
//is possible.  The GetBody func is used to see if the command has a body
//that needs to be displayed in the browser.  This is used by the client
//side to know when it should copy data to the display.
//
//Note that this structure is _never_ needed on the server side (inside 
//seven5) because it always returns a structure for marshalling.  That
//structure may indicate that there is a problem but that isn't tested
//until the value reaches the client (calling) side.
//Public so others can write commands.
type CommandReturn struct {
	Unmarshalled func() interface{}
	ErrorTest func(interface{}) bool
	GetBody func(interface{}) string
}

// SimpleErrorReturn is suitable as a return type for commands that just signal
// an error or not. Public so others can write commands.
type SimpleErrorReturn struct {
	Error bool
}
//BodyReturn is suitable as a return type for commands that needs to write
//data to the browser. Public so others can write commands.
type BodyReturn struct {
	Error bool
	Body string
}

//simpleError tests a SimpleErrorReturn for its Error field
func simpleErrorTest(v interface{}) bool {
	return v.(*SimpleErrorReturn).Error
}
//bodyError tests a BodyReturn for its Error field
func bodyErrorTest(v interface{}) bool {
	return v.(*BodyReturn).Error
}
//bodyGetBody retreives the body from a bodyReturn
func bodyGetBody(v interface{}) string {
	return v.(*BodyReturn).Body
}

//CommandArgPair represents a single argument to a command.  The Generator is
// used on the _client_ side to generate the value to be passed to the seven5
//implementation. The result is tested to see if it is a string and if not
//it is encoded into json.   If the result *is* a string then Unmarshalled
//should be the nil func as there is no encoding.
//Unmarshalled is used on the server side to get an instance of the correct
//type that we can unmarshal the parameter into.
//Public so others can write commands.
type CommandArgPair struct {
	Unmarshalled func() interface{}
	Generator func(ClientSideCapability,util.SimpleLogger) (interface{}, error)
}

