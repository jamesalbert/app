package cef

import (
	"os"

	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/internal/bridge"
	"github.com/murlokswarm/app/internal/core"
	"github.com/murlokswarm/app/internal/logs"
)

var (
	driver *Driver
)

func init() {
	logger := logs.ToWriter(os.Stderr)
	app.Logger = logs.WithColoredPrompt(logger)
}

// Driver implements the app.Driver interface.
type Driver struct {
	core.Driver

	factory     *app.Factory
	elems       *core.ElemDB
	platformRPC bridge.PlatformRPC
	goRPC       bridge.GoRPC
	uichan      chan func()
}

func (d *Driver) Run(f *app.Factory) error {
	d.factory = f
	d.elems = core.NewElemDB()
	d.uichan = make(chan func(), 256)

	driver = d
	return nil
}

// CallOnUIGoroutine satisfies the app.Driver interface.
func (d *Driver) CallOnUIGoroutine(f func()) {
	d.uichan <- f
}
