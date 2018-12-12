package sqlx

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogap/config"
	"github.com/gogap/mdv"
	"github.com/jmoiron/sqlx"
	//  _ "github.com/go-sql-driver/mysql"
)

var _ mdv.DataFiller = (*SQLXDataFiller)(nil)

type DbConn struct {
	Name     string
	DSN      string
	Driver   string
	Filename string
	UpdateAt time.Time
}

type ModelConfig struct {
	Filename string
	Name     string
	SQL      string
	ConnName string
	UpdateAt time.Time
}

type SQLXDataFiller struct {
	DbConns   map[string]*DbConn
	modelSQLs map[string]*ModelConfig
	connExpir time.Duration
	sqlExpir  time.Duration
	connsDir  string
	modelsDir string
}

func init() {
	mdv.RegisterDataFiller("sqlx", NewSQLXDataFiller)
}

func NewSQLXDataFiller(conf config.Configuration) (v mdv.DataFiller, err error) {

	var dataFiller = SQLXDataFiller{
		DbConns:   make(map[string]*DbConn),
		modelSQLs: make(map[string]*ModelConfig),
		connExpir: time.Minute,
		sqlExpir:  time.Minute,
	}

	connsDir := conf.GetString("connections.dir")
	if len(connsDir) == 0 {
		err = fmt.Errorf("the SQLXDataFiller's config of connections.dir is empty")
		return
	}

	dataFiller.connsDir = connsDir

	connFiles, err := ioutil.ReadDir(connsDir)
	if err != nil {
		return
	}

	for _, file := range connFiles {

		if file.IsDir() {
			continue
		}

		ext := filepath.Ext(file.Name())

		if ext != ".conf" {
			continue
		}

		filename := filepath.Join(connsDir, file.Name())

		err = dataFiller.loadConn(filename)
		if err != nil {
			return
		}
	}

	sqlsDir := conf.GetString("sqls.dir")
	if len(sqlsDir) == 0 {
		err = fmt.Errorf("the SQLXDataFiller's config of sqls.dir is empty")
		return
	}

	dataFiller.modelsDir = sqlsDir

	sqlFiles, err := ioutil.ReadDir(sqlsDir)
	if err != nil {
		return
	}

	for _, file := range sqlFiles {

		if file.IsDir() {
			continue
		}

		ext := filepath.Ext(file.Name())

		if ext != ".sql" {
			continue
		}

		filename := filepath.Join(sqlsDir, file.Name())

		err = dataFiller.loadModels(filename)
		if err != nil {
			return
		}
	}

	dataFiller.connExpir = conf.GetTimeDuration("conn-expir", time.Minute)
	dataFiller.sqlExpir = conf.GetTimeDuration("sql-expir", time.Minute)

	v = &dataFiller

	return
}

func (p *SQLXDataFiller) Fill(modelName string, modProducer mdv.ModelProducer, args map[string]interface{}) (value interface{}, err error) {

	if modProducer == nil {
		err = fmt.Errorf("model producer is nil")
		return
	}

	modelConf, exist := p.modelSQLs[modelName]
	if !exist {

		loadPath := filepath.Join(p.modelsDir, modelName+".conf")

		err = p.loadModels(loadPath)
		if err != nil {
			return
		}

		modelConf, exist = p.modelSQLs[modelName]
		if !exist {
			err = fmt.Errorf("the model name of %s' sql config not exist, from: %s", modelName, loadPath)
			return
		}
	}

	conn, exist := p.DbConns[modelConf.ConnName]

	if !exist {

		loadPath := filepath.Join(p.connsDir, modelConf.ConnName+".conf")
		err = p.loadConn(loadPath)
		if err != nil {
			return
		}

		conn, exist = p.DbConns[modelConf.ConnName]
		if !exist {
			err = fmt.Errorf("the connection name of %s not exist, current model is: %s, from: %s", modelConf.ConnName, modelName, loadPath)
			return
		}
	}

	now := time.Now()

	if p.connExpir > 0 {
		if now.Sub(conn.UpdateAt) > p.connExpir {
			err = conn.Reload()
			if err != nil {
				return
			}
		}
	}

	if p.sqlExpir > 0 {
		if now.Sub(modelConf.UpdateAt) > p.sqlExpir {
			err = modelConf.Reload()
			if err != nil {
				return
			}
		}
	}

	db, err := sqlx.Connect(conn.Driver, conn.DSN)
	if err != nil {
		return
	}

	defer db.Close()

	rows, err := db.NamedQuery(modelConf.SQL, args)

	if err != nil {
		return
	}

	var list []interface{}

	for rows.Next() {
		var newMod interface{}
		newMod, err = modProducer()

		if err != nil {
			return
		}
		err = rows.StructScan(newMod)
		if err != nil {
			return
		}
		list = append(list, newMod)
	}

	value = list

	return
}

func (p *DbConn) Reload() (err error) {
	conf := config.NewConfig(config.ConfigFile(p.Filename))

	if conf.IsEmpty() {
		err = fmt.Errorf("conn config file of %s is empty", p.Filename)
		return
	}

	name := filepath.Base(p.Filename)
	ext := filepath.Ext(name)
	connName := strings.TrimSuffix(name, ext)

	dsn := conf.GetString("dsn")
	if len(dsn) == 0 {
		err = fmt.Errorf("the SQLXDataFiller's config of connections.%s.dsn is empty", connName)
		return
	}

	driver := conf.GetString("driver")
	if len(driver) == 0 {
		err = fmt.Errorf("the SQLXDataFiller's config of connections.%s.driver is empty", connName)
		return
	}

	p.Name = connName
	p.DSN = dsn
	p.Driver = driver
	p.UpdateAt = time.Now()

	return
}

func (p *ModelConfig) Reload() (err error) {
	conf := config.NewConfig(config.ConfigFile(p.Filename))

	if conf.IsEmpty() {
		err = fmt.Errorf("model config file of %s is empty", p.Filename)
		return
	}

	name := filepath.Base(p.Filename)
	ext := filepath.Ext(name)
	modelName := strings.TrimSuffix(name, ext)

	sql := conf.GetString("sql")
	if len(sql) == 0 {
		err = fmt.Errorf("the SQLXDataFiller's config of model.%s.sql is empty", modelName)
		return
	}

	conn := conf.GetString("conn")
	if len(conn) == 0 {
		err = fmt.Errorf("the SQLXDataFiller's config of model.%s.conn is empty", modelName)
		return
	}

	p.Name = modelName
	p.ConnName = conn
	p.SQL = sql
	p.UpdateAt = time.Now()

	return
}

func (p *SQLXDataFiller) loadConn(filename string) (err error) {

	name := filepath.Base(filename)
	ext := filepath.Ext(name)
	connName := strings.TrimSuffix(name, ext)

	conn := &DbConn{
		Name:     connName,
		Filename: filename,
	}

	err = conn.Reload()
	if err != nil {
		return
	}

	p.DbConns[conn.Name] = conn

	return
}

func (p *SQLXDataFiller) loadModels(filename string) (err error) {

	name := filepath.Base(filename)
	ext := filepath.Ext(name)
	modelName := strings.TrimSuffix(name, ext)

	model := &ModelConfig{
		Name:     modelName,
		Filename: filename,
	}

	err = model.Reload()
	if err != nil {
		return
	}

	p.modelSQLs[model.Name] = model

	return
}
