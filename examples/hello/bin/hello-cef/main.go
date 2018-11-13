package main

import (
	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/drivers/cef"
)

func main() {
	app.Run(&cef.Driver{})
}
