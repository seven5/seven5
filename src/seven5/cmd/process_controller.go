package cmd

import (
	"seven5/util"
)

//
//  To Be Implemented: Process A Controller
//
var ProcessController = &CommandDecl{
	Arg: []*CommandArgPair{
		ClientSideWd, //root of the user project
	},
	Ret: BuiltinSimpleReturn,
	Impl: defaultProcessController,
}


func defaultProcessController(log util.SimpleLogger, v...interface{}) interface{} {

	return &SimpleErrorReturn{Error: true}
}
