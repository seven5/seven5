package hellorest

import (
	"seven5"
)

//This will be detected and filled as a parameter to your start func.
//Field names should be upper case because otherwise they can't be
//json encoded.
type UserConfig struct {
	Foo int
	Bar string
	Baz float64
}

//Start is called by seven5 to intialize your app. The second parameter
//will be of type *UserConfigStruct if you have one.
func start(app *seven5.ApplicationConfig, user *UserConfig) {
}
