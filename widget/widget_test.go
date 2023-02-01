package widget

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaoapp/gou/process"
	"github.com/yaoapp/kun/any"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/share"
)

func TestLoad(t *testing.T) {
	// runtime.Load(config.Conf)
	share.DBConnect(config.Conf.DB) // 创建数据库连接
	Load(config.Conf)
	v, err := process.New("widgets.dyform.Save", "pad", "pay").Exec()
	if err != nil {
		t.Fatal(err)
	}

	res := any.Of(v).Map()
	assert.Equal(t, "pad", res.Get("instance"))
	assert.Equal(t, "pay", res.Get("payload"))
}
