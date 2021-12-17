package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"io"
	"os"

	"example.com/helloworld/app/hw/service/internal/conf"
	xlog "example.com/helloworld/pkg/log"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/natefinch/lumberjack"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
	)
}

func main() {
	flag.Parse()
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var lc conf.Log
	if err := c.Scan(&lc); err != nil {
		panic(err)
	}
	ll, _ := logrus.ParseLevel(lc.Level)
	lo := io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename:   lc.Dir,
		MaxSize:    int(lc.MaxSize),    // megabytes
		MaxBackups: int(lc.MaxBackups), // numbers
		MaxAge:     int(lc.MaxAge),     // days
		Compress:   lc.Compress,        // disabled by default
	})
	formatter := func() logrus.Formatter {
		if lc.JsonFormatter {
			return &logrus.JSONFormatter{}
		}
		return &logrus.TextFormatter{}
	}()
	logger := xlog.NewLogrusLogger(xlog.Level(ll), xlog.Output(lo), xlog.Formatter(formatter))
	logger = log.With(
		logger,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"span_id", tracing.SpanID(),
		"trace_id", tracing.TraceID(),
	)

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := initApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
