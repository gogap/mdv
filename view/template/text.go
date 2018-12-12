package template

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/gogap/config"
	"github.com/gogap/mdv"
)

var _ mdv.Viewer = (*TextView)(nil)

type TextView struct {
	tmplDir string
}

func init() {
	mdv.RegisterViewer("text", NewTextView)
}

func NewTextView(conf config.Configuration) (v mdv.Viewer, err error) {

	tmplDir := conf.GetString("dir")

	if len(tmplDir) == 0 {
		err = fmt.Errorf("the config of template dir is empty")
		return
	}

	v = &TextView{
		tmplDir: tmplDir,
	}

	return
}

func (p *TextView) Render(modelName, viewName string, model interface{}, args map[string]interface{}, w io.Writer) (err error) {

	filename := filepath.Join(p.tmplDir, viewName, modelName+".tmpl")

	_, err = os.Stat(filename)

	if err != nil {
		filename = filepath.Join(p.tmplDir, viewName, "default.tmpl")

		_, err = os.Stat(filename)

		if err != nil {
			return
		}
	}

	name := filepath.Base(filename)
	tmpl, err := template.New(name).Funcs(sprig.TxtFuncMap()).ParseFiles(filename)
	if err != nil {
		return
	}

	metafile := filepath.Join(p.tmplDir, viewName, modelName+".json")

	var metadata map[string]interface{}

	if data, e := ioutil.ReadFile(metafile); e == nil {
		err = json.Unmarshal(data, &metadata)
		if err != nil {
			return
		}
	}

	err = tmpl.Execute(w, map[string]interface{}{"model": model, "args": args, "metadata": metadata})
	if err != nil {
		return
	}

	return
}
