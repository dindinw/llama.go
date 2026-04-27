package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/Qitmeer/llama.go/config"
	"github.com/Qitmeer/llama.go/runner"
	"github.com/Qitmeer/llama.go/server/middleware"
	"github.com/Qitmeer/llama.go/server/routes"
	"github.com/Qitmeer/llama.go/version"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
)

type Service struct {
	ctx *cli.Context
	cfg *config.Config

	addr net.Addr
	srvr *http.Server

	wg sync.WaitGroup

	api *routes.API

	runnerSer *runner.Service
}

func New(ctx *cli.Context, cfg *config.Config) *Service {
	log.Info("New Server ...")
	runnerSer := runner.New(ctx, cfg)
	ser := Service{ctx: ctx, cfg: cfg, api: routes.New(cfg, runnerSer), runnerSer: runnerSer}
	return &ser
}

func (s *Service) Start() error {
	log.Info("Start Server...")
	err := s.runnerSer.Start()
	if err != nil {
		return err
	}
	err = s.api.Start()
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", s.cfg.HostURL().Host)
	if err != nil {
		return err
	}
	s.addr = ln.Addr()

	err = s.GenerateRoutes()
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Listening on %s (%s) ", ln.Addr(), version.Details()))
	s.srvr = &http.Server{
		Handler: nil,
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		err = s.srvr.Serve(ln)
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error(err.Error())
		}
	}()

	return nil
}

func (s *Service) GenerateRoutes() error {
	r := gin.Default()

	r.Use(middleware.Security())
	r.Use(middleware.CORS(s.cfg.AllowedOrigins()))
	r.Use(middleware.AllowedHosts(s.addr))

	r.HandleMethodNotAllowed = true

	s.api.Setup(r)

	http.Handle("/", r)
	return nil
}

func (s *Service) Stop() error {
	log.Info("Stop Server...")

	var err error
	if s.srvr != nil {
		err = s.srvr.Close()
	}
	s.wg.Wait()

	if s.runnerSer != nil {
		err = s.runnerSer.Stop()
	}
	return err
}
