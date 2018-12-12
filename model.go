package mdv

import (
	"fmt"
	"github.com/gogap/config"
)

type NewModelerFunc func(conf config.Configuration) (Modeler, error)

var (
	modelers = make(map[string]NewModelerFunc)
)

type Modeler interface {
	New(modelName string) (mod interface{}, exist bool)
}

type ModelProducer func() (interface{}, error)

func RegisterModeler(driverName string, fn NewModelerFunc) {

	if fn == nil {
		panic("NewModelerFunc is nil")
	}

	_, exist := modelers[driverName]
	if exist {
		panic("modelers's driver already registerd: " + driverName)
	}

	modelers[driverName] = fn
}

func NewModeler(driverName string, conf config.Configuration) (modeler Modeler, err error) {

	fn, exist := modelers[driverName]

	if !exist {
		err = fmt.Errorf("the model driver of %s not exist", driverName)
		return
	}

	modeler, err = fn(conf)

	return
}
