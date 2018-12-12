package mdv

import (
	"fmt"
	"io"

	"github.com/gogap/config"
)

type NewViewerFunc func(conf config.Configuration) (Viewer, error)

var (
	viewers = make(map[string]NewViewerFunc)
)

type Viewer interface {
	Render(modelName, viewName string, model interface{}, args map[string]interface{}, rw io.Writer) error
}

func RegisterViewer(driverName string, fn NewViewerFunc) {

	if fn == nil {
		panic("NewViewerFunc is nil")
	}

	_, exist := viewers[driverName]
	if exist {
		panic("viewers's driver already registerd: " + driverName)
	}

	viewers[driverName] = fn
}

func NewViewer(driverName string, conf config.Configuration) (viewer Viewer, err error) {

	fn, exist := viewers[driverName]

	if !exist {
		err = fmt.Errorf("the view driver of %s not exist", driverName)
		return
	}

	viewer, err = fn(conf)

	return
}
