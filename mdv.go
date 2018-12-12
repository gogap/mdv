package mdv

import (
	"bytes"
	"fmt"
	"github.com/gogap/config"
)

type MDV struct {
	Config config.Configuration

	model Modeler
	data  DataFiller
	view  Viewer
}

func New(configFile string) (m *MDV, err error) {

	conf := config.NewConfig(config.ConfigFile(configFile))

	tmpMDV := &MDV{
		Config: conf,
	}

	err = tmpMDV.init()
	if err != nil {
		return
	}

	m = tmpMDV

	return
}

func (p *MDV) init() (err error) {

	dataConf := p.Config.GetConfig("data")
	viewConf := p.Config.GetConfig("view")
	modelConf := p.Config.GetConfig("model")

	dataDriver := dataConf.GetString("driver")
	viewDriver := viewConf.GetString("driver")
	modelDriver := modelConf.GetString("driver")

	if len(dataDriver) == 0 || len(viewDriver) == 0 || len(modelDriver) == 0 {
		err = fmt.Errorf("(DataDriver and ModelDriver and ViewDriver) could not be empty")
		return
	}

	data, err := NewDataFiller(dataDriver, dataConf.GetConfig("options"))
	if err != nil {
		return
	}

	view, err := NewViewer(viewDriver, viewConf.GetConfig("options"))
	if err != nil {
		return
	}

	model, err := NewModeler(modelDriver, modelConf.GetConfig("options"))
	if err != nil {
		return
	}

	p.model = model
	p.view = view
	p.data = data

	return
}

func (p *MDV) Execute(modelName string, viewName string, dataArgs map[string]interface{}, filters ...DataFilter) (data []byte, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	var producer ModelProducer = func() (mod interface{}, err error) {
		m, exist := p.model.New(modelName)

		if !exist {
			fmt.Errorf("model of %s not exist", modelName)
			return
		}

		mod = m
		return
	}

	value, err := p.data.Fill(modelName, producer, dataArgs)
	if err != nil {
		return
	}

	if len(filters) > 0 {
		for i := 0; i < len(filters); i++ {
			err = filters[i](modelName, value, dataArgs)
			if err != nil {
				return
			}
		}
	}

	buf := bytes.NewBuffer(nil)

	err = p.view.Render(modelName, viewName, value, dataArgs, buf)

	if err != nil {
		return
	}

	data = buf.Bytes()

	return
}
