package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Qitmeer/llama.go/config"
	"github.com/Qitmeer/llama.go/wrapper"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

type Service struct {
	ctx     *cli.Context
	cfg     *config.Config
	running bool
}

func New(ctx *cli.Context, cfg *config.Config) *Service {
	log.Info("New Runner ...")
	ser := Service{ctx: ctx, cfg: cfg, running: false}
	return &ser
}

func (s *Service) Start() error {
	if s.IsRunning() {
		return errors.New("Already Running")
	}
	log.Info("Start Runner...")

	errCh := make(chan error, 1)
	go func() {
		errCh <- wrapper.LlamaStart(s.cfg)
	}()

	timeout := time.After(15 * time.Minute)
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case err := <-errCh:
			if err != nil {
				log.Error(err.Error())
				return fmt.Errorf("llama core failed to start: %w", err)
			}
			return errors.New("llama core stopped before becoming ready")
		case <-tick.C:
			if wrapper.IsLlamaRunning() {
				s.running = true
				log.Info("Started llama core (inference loop ready)")
				go func() {
					err := <-errCh
					if err != nil {
						log.Error("llama core stopped", "error", err)
					} else {
						log.Info("llama core stopped")
					}
					s.running = false
				}()
				return nil
			}
		case <-timeout:
			return errors.New("timeout waiting for llama core to become ready")
		}
	}
}

func (s *Service) Stop() error {
	if !s.IsRunning() {
		return errors.New("Not running")
	}
	log.Info("Stop Runner...")
	err := wrapper.LlamaStop()
	if err != nil {
		log.Error(err.Error())
	}
	s.running = false
	return nil
}

func (s *Service) IsRunning() bool {
	return s.running
}

func (s *Service) Generate(id int, model string, prompt string, stream bool) error {
	type body struct {
		Model  string `json:"model,omitempty"`
		Prompt string `json:"prompt"`
		Stream bool   `json:"stream"`
	}
	b, err := json.Marshal(body{Model: model, Prompt: prompt, Stream: stream})
	if err != nil {
		return err
	}
	return wrapper.LlamaGenerate(id, string(b))
}

func (s *Service) Chat(id int, model string, jsStr string) error {
	payload := jsStr
	if model != "" {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(jsStr), &obj); err != nil {
			return err
		}
		obj["model"] = model
		b, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		payload = string(b)
	}
	return wrapper.LlamaChat(id, payload)
}
