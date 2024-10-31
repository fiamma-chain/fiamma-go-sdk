package context

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/fiamma-chain/fiamma-go-sdk/errors"
	"github.com/fiamma-chain/fiamma-go-sdk/log"
	"github.com/fiamma-chain/fiamma-go-sdk/utils"
)

// Context of service
type Context interface {
	// ConfFile returns config file from data.
	ConfFile() string

	// Log returns logger interface.
	Log() *log.Logger

	// Wait waits until exit, receiving SIGTERM and SIGINT signals.
	Wait()
	// WaitChan returns wait channel.
	WaitChan() <-chan os.Signal

	// LoadCustomConfig loads custom config.
	// If 'files' is empty, will load config from default path,
	// else the first file path will be used to load config from.
	LoadCustomConfig(cfg interface{}, files ...string) error

	Done()
}

type ctx struct {
	sync.Map // global cache
	log      *log.Logger
	sig      chan os.Signal
}

// NewContext creates a new context
func NewContext(confFile string) Context {
	c := &ctx{}
	c.Store(KeyConfFile, confFile)

	var lfs []log.Field
	c.log = log.With(lfs...)
	c.log.Info("to load config file", log.Any("file", c.ConfFile()))

	sc := &SystemConfig{}
	err := c.LoadCustomConfig(sc)
	if err != nil {
		c.log.Error("failed to load system config, to use default config", log.Error(err))
	}

	_log, err := log.Init(sc.Logger, lfs...)
	if err != nil {
		c.log.Error("failed to init logger", log.Error(err))
	}
	c.log = _log
	c.log.Debug("context is created", log.Any("file", confFile), log.Any("conf", sc))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	c.sig = sig
	return c
}

func (c *ctx) ConfFile() string {
	v, ok := c.Load(KeyConfFile)
	if !ok {
		return ""
	}
	return v.(string)
}

func (c *ctx) Log() *log.Logger {
	return c.log
}

func (c *ctx) Wait() {
	<-c.sig
}

func (c *ctx) WaitChan() <-chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	return sig
}

func (c *ctx) LoadCustomConfig(cfg interface{}, files ...string) error {
	f := c.ConfFile()
	if len(files) > 0 && len(files[0]) > 0 {
		f = files[0]
	}
	if utils.FileExists(f) {
		return errors.Trace(utils.LoadYAML(f, cfg))
	}
	return errors.Trace(utils.UnmarshalYAML(nil, cfg))
}

func (c *ctx) Done() {
	c.sig <- syscall.SIGKILL
}
