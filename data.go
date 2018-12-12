package mdv

import (
	"fmt"

	"github.com/gogap/config"
)

type DataFilter func(modelName string, v interface{}, dataArgs map[string]interface{}) (err error)

type NewDataFillerFunc func(conf config.Configuration) (DataFiller, error)

var (
	dataFillers = make(map[string]NewDataFillerFunc)
)

type DataFiller interface {
	Fill(modelName string, modProducer ModelProducer, args map[string]interface{}) (value interface{}, err error)
}

func RegisterDataFiller(driverName string, fn NewDataFillerFunc) {

	if fn == nil {
		panic("NewDataFillerFunc is nil")
	}

	_, exist := dataFillers[driverName]
	if exist {
		panic("dataFillers's driver already registerd: " + driverName)
	}

	dataFillers[driverName] = fn
}

func NewDataFiller(driverName string, conf config.Configuration) (filler DataFiller, err error) {

	fn, exist := dataFillers[driverName]

	if !exist {
		err = fmt.Errorf("the data filler driver of %s not exist", driverName)
		return
	}

	filler, err = fn(conf)

	return
}
