package client

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	//	_ "honnef.co/go/js/console"
)

type Application interface {
	Start()
}

func Main(app Application) {
	jquery.NewJQuery(js.Global.Get("document")).Ready(func() {
		app.Start()
	})
}
