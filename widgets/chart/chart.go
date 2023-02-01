package chart

import (
	"fmt"
	"path/filepath"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/yaoapp/gou/process"
	"github.com/yaoapp/kun/exception"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/share"
	"github.com/yaoapp/yao/widgets/component"
	"github.com/yaoapp/yao/widgets/field"
)

//
// API:
//   GET  /api/__yao/chart/:id/setting  					-> Default process: yao.chart.Xgen
//   GET  /api/__yao/chart/:id/data 						-> Default process: yao.chart.Data $param.id :query
//   GET  /api/__yao/chart/:id/component/:xpath/:method  	-> Default process: yao.chart.Component $param.id $param.xpath $param.method :query
//
// Process:
// 	 yao.form.Setting Return the App DSL
// 	 yao.form.Xgen Return the Xgen setting
//   yao.form.Data Return the query data
//   yao.form.Component Return the result defined in props.xProps
//
// Hook:
//   before:data
//   after:data
//

// Charts the loaded chart widgets
var Charts map[string]*DSL = map[string]*DSL{}

// New create a new DSL
func New(id string) *DSL {
	return &DSL{
		ID:     id,
		Fields: &FieldsDSL{Chart: field.Columns{}, Filter: field.Filters{}},
		CProps: field.CloudProps{},
		Config: map[string]interface{}{},
	}
}

// LoadAndExport load table
func LoadAndExport(cfg config.Config) error {
	err := Load(cfg)
	if err != nil {
		return err
	}
	return Export()
}

// Load load task
func Load(cfg config.Config) error {
	var root = filepath.Join(cfg.Root, "charts")
	return LoadFrom(root, "")
}

// LoadFrom load from dir
func LoadFrom(dir string, prefix string) error {

	if share.DirNotExists(dir) {
		return fmt.Errorf("%s does not exists", dir)
	}

	messages := []string{}
	err := share.Walk(dir, ".json", func(root, filename string) {
		id := prefix + share.ID(root, filename)
		data := share.ReadFile(filename)
		dsl := New(id)
		err := jsoniter.Unmarshal(data, dsl)
		if err != nil {
			messages = append(messages, fmt.Sprintf("[%s] %s", id, err.Error()))
			return
		}

		if dsl.Action == nil {
			dsl.Action = &ActionDSL{}
		}
		dsl.Action.SetDefaultProcess()

		if dsl.Layout == nil {
			dsl.Layout = &LayoutDSL{}
		}

		// mapping
		err = dsl.mapping()
		if err != nil {
			messages = append(messages, fmt.Sprintf("[%s] Mapping %s", id, err.Error()))
			return
		}

		// Validate
		err = dsl.Validate()
		if err != nil {
			messages = append(messages, fmt.Sprintf("[%s] %s", id, err.Error()))
			return
		}

		Charts[id] = dsl
	})

	if len(messages) > 0 {
		return fmt.Errorf(strings.Join(messages, ";"))
	}

	return err
}

// Get chart via process or id
func Get(chart interface{}) (*DSL, error) {
	id := ""
	switch chart.(type) {
	case string:
		id = chart.(string)
	case *process.Process:
		id = chart.(*process.Process).ArgsString(0)
	default:
		return nil, fmt.Errorf("%v type does not support", chart)
	}

	t, has := Charts[id]
	if !has {
		return nil, fmt.Errorf("%s does not exist", id)
	}
	return t, nil
}

// MustGet Get chart via process or id thow error
func MustGet(chart interface{}) *DSL {
	t, err := Get(chart)
	if err != nil {
		exception.New(err.Error(), 400).Throw()
	}
	return t
}

// Xgen trans to xgen setting
func (dsl *DSL) Xgen(data map[string]interface{}, excludes map[string]bool) (map[string]interface{}, error) {

	layout, err := dsl.Layout.Xgen(data, excludes, dsl.Mapping)
	if err != nil {
		return nil, err
	}

	fields, err := dsl.Fields.Xgen(layout)
	if err != nil {
		return nil, err
	}

	// full width default value
	if _, has := dsl.Config["full"]; !has {
		dsl.Config["full"] = true
	}

	setting := map[string]interface{}{}
	bytes, err := jsoniter.Marshal(layout)
	if err != nil {
		return nil, err
	}

	err = jsoniter.Unmarshal(bytes, &setting)
	if err != nil {
		return nil, err
	}

	setting["name"] = dsl.Name
	setting["fields"] = fields
	setting["config"] = dsl.Config
	for _, cProp := range dsl.CProps {
		err := cProp.Replace(setting, func(cProp component.CloudPropsDSL) interface{} {
			return map[string]interface{}{
				"api":    fmt.Sprintf("/api/__yao/chart/%s%s", dsl.ID, cProp.Path()),
				"params": cProp.Query,
			}
		})
		if err != nil {
			return nil, err
		}
	}

	return setting, nil
}

// Actions get the chart actions
func (dsl *DSL) Actions() []component.ActionsExport {

	res := []component.ActionsExport{}

	// layout.operation.actions
	if dsl.Layout != nil &&
		dsl.Layout.Operation != nil &&
		dsl.Layout.Operation.Actions != nil &&
		len(dsl.Layout.Operation.Actions) > 0 {

		res = append(res, component.ActionsExport{
			Type:    "operation",
			Xpath:   "layout.operation.actions",
			Actions: dsl.Layout.Operation.Actions,
		})
	}

	// layout.filter.actions
	if dsl.Layout != nil &&
		dsl.Layout.Filter != nil &&
		dsl.Layout.Filter.Actions != nil &&
		len(dsl.Layout.Filter.Actions) > 0 {

		res = append(res, component.ActionsExport{
			Type:    "filter",
			Xpath:   "layout.filter.actions",
			Actions: dsl.Layout.Filter.Actions,
		})
	}
	return res
}
