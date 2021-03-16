package server

import (
	"github.com/gin-gonic/gin"
	"github.com/spongeprojects/kubebigbrother/pkg/log"
)

type Options struct {
	Version string

	Env string

	Addr     string
	GinDebug bool
}

type App struct {
	Version string

	Addr string

	Router *gin.Engine
}

func SetupApp(options *Options) (*App, error) {
	if options == nil {
		options = &Options{}
	}

	app := &App{}
	app.Addr = options.Addr

	if options.GinDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(
		gin.LoggerWithConfig(gin.LoggerConfig{
			Output:    log.Logger.Out,
			SkipPaths: []string{"/healthz"},
		}),
		gin.Recovery(),
	)

	r.GET("/", app.Index)
	r.Any("/healthz", app.Healthz)

	r.HandleMethodNotAllowed = true

	app.Router = r

	return app, nil
}

func (app *App) Serve() error {
	log.Infof("serving on %s", app.Addr)
	return app.Router.Run(app.Addr)
}
