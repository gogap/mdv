package main

import (
	"fmt"
	"github.com/gogap/mdv"
)

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/gogap/mdv/data/sqlx"
	_ "github.com/gogap/mdv/model/dmod"
	_ "github.com/gogap/mdv/view/template"
)

func main() {

	var err error

	defer func() {
		if err != nil {
			fmt.Println(err)
		}
	}()

	m, err := mdv.New("mdv.conf")
	if err != nil {
		return
	}

	data, err := m.Execute("User", "bar", map[string]interface{}{"id": 1})

	if err != nil {
		return
	}

	fmt.Println(string(data))
}
