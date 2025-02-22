package service

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yaoapp/gou/api"
	"github.com/yaoapp/gou/server/http"
	"github.com/yaoapp/yao/config"
	"github.com/yaoapp/yao/share"
)

// Start the yao service
func Start(cfg config.Config) (*http.Server, error) {

	if cfg.AllowFrom == nil {
		cfg.AllowFrom = []string{}
	}

	err := prepare()
	if err != nil {
		return nil, err
	}

	router := gin.New()
	router.Use(gin.Logger())
	api.SetGuards(Guards)
	api.SetRoutes(router, "/api", cfg.AllowFrom...)

	srv := http.New(router, http.Option{
		Host:    cfg.Host,
		Port:    cfg.Port,
		Root:    "/api",
		Allows:  cfg.AllowFrom,
		Timeout: 5 * time.Second,
	}).With(Middlewares...)

	go func() {
		err = srv.Start()
	}()

	return srv, nil
}

// Stop the yao service
func Stop(srv *http.Server) error {
	err := srv.Stop()
	if err != nil {
		return err
	}
	<-srv.Event()
	return nil
}

func prepare() error {

	// Session server
	err := share.SessionStart()
	if err != nil {
		return err
	}

	err = SetupStatic()
	if err != nil {
		return err
	}

	return nil
}
