package dmod

import (
	"time"

	"github.com/gogap/config"
	"github.com/gogap/dmod"
	"github.com/gogap/mdv"
)

var _ mdv.Modeler = (*DModModel)(nil)

type DModModel struct {
	dMod      *dmod.Models
	updatedAt time.Time
	expire    time.Duration
	dir       string
}

func init() {
	mdv.RegisterModeler("dmod", NewDModModel)
}

func NewDModModel(conf config.Configuration) (v mdv.Modeler, err error) {

	models, err := dmod.NewModels()
	if err != nil {
		return
	}

	dm := &DModModel{
		dMod:      models,
		updatedAt: time.Now(),
	}

	dir := conf.GetString("dir")

	if len(dir) == 0 {
		v = dm
		return
	}

	err = models.LoadFromDir(dir)

	if err != nil {
		return
	}

	expireDur := conf.GetTimeDuration("expire", time.Minute)

	dm.dir = dir
	dm.expire = expireDur

	v = dm

	return
}

func (p *DModModel) New(modelName string) (mod interface{}, exist bool) {

	m, exist := p.dMod.GetModel(modelName)

	if !exist {
		return
	}

	mod = m.New()

	if p.shouldUpdate() {
		go p.reloadModels()
	}

	return
}

func (p *DModModel) shouldUpdate() bool {
	if p.expire > 0 {

		exp := p.expire
		if exp < time.Second*10 {
			exp = time.Second * 10
		}

		if time.Now().Sub(p.updatedAt) > exp {
			return true
		}
	}

	return false
}

func (p *DModModel) reloadModels() (err error) {
	var models *dmod.Models
	models, err = dmod.NewModels()
	if err != nil {
		return
	}
	err = models.LoadFromDir(p.dir)
	if err != nil {
		return
	}

	p.updatedAt = time.Now()
	p.dMod = models

	return
}
