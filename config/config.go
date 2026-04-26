// Copyright (c) 2017-2025 The qitmeer developers

package config

import (
	"fmt"
	"math"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Qitmeer/llama.go/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

const (
	// ---------------------------------------------------------------------
	// Defaults in this file are intentionally aligned with upstream llama.cpp
	// (struct `common_params`). Keeping them in sync ensures that developers
	// familiar with llama.cpp see consistent behavior in llama.go, and avoids
	// subtle bugs such as prompt truncation caused by divergent defaults.
	//
	// Upstream reference :
	//   - llama.cpp `common/common.h`, field: `common_params`
	//   - URL   : https://github.com/Qitmeer/llama.cpp/blob/29db96f3256f839cb9bd6f72056016dcae2214ef/common/common.h#L246
	//
	// If upstream changes the defaults, please update the corresponding
	// constants here and keep this comment’s commit hash in sync.
	// ---------------------------------------------------------------------

	defaultLogLevel = "info"

	// defaultNPredict controls the maximum number of tokens to generate.
	// Semantics:
	//   -1 : unlimited (generation is only bounded by stop conditions)
	//   -2 : generate until the context window is filled (supported by llama.cpp)
	//
	// Rationale:
	//   We keep the default at -1 to mirror llama.cpp’s `common_params::n_predict`
	//   in order to avoid reserving a fixed output budget (e.g., 512 tokens)
	//   that would reduce the effective prompt capacity (`n_ctx - n_predict`)
	//   and lead to surprising prompt truncation when clients omit num_predict.
	defaultNPredict = -1

	DefaultHost        = "127.0.0.1:8081"
	DefaultPort        = "8081"
	DefaultModelDir    = "./data/models"
	DefaultContextSize = 4096

	DefaultNGpuLayers = -1

	EXT     = ".gguf" // TODO:We will soon release our better format
	EXT_BIN = ".bin"
)

var (
	defaultHomeDir     = "."
	defaultSwaggerFile = filepath.Join(defaultHomeDir, "swagger.json")
)

var (
	Conf = &Config{}

	LogLevel = &cli.StringFlag{
		Name:        "log-level",
		Aliases:     []string{"l"},
		Usage:       "Logging level {trace, debug, info, warn, error}",
		Value:       defaultLogLevel,
		Destination: &Conf.LogLevel,
	}

	Model = &cli.StringFlag{
		Name:        "model",
		Aliases:     []string{"m"},
		Usage:       "The name of the model file located in the 'model-dir' repository path or absolute path of resource file",
		Destination: &Conf.Model,
	}

	ModelDir = &cli.StringFlag{
		Name:        "model-dir",
		Aliases:     []string{"md"},
		Usage:       "Path for storing model files",
		Value:       DefaultModelDir,
		Destination: &Conf.ModelDir,
		EnvVars:     []string{"LLAMAGO_MODEL_DIR"},
	}

	CtxSize = &cli.IntFlag{
		Name:        "ctx-size",
		Aliases:     []string{"c"},
		Usage:       "Set the size of the prompt context. The default is 4096, but if a LLaMA model was built with a longer context, increasing this value will provide better results for longer input/inference",
		Value:       DefaultContextSize,
		Destination: &Conf.CtxSize,
	}

	Prompt = &cli.StringFlag{
		Name:        "prompt",
		Aliases:     []string{"p"},
		Usage:       "Provide a prompt directly as a command-line option.",
		Destination: &Conf.Prompt,
	}

	NGpuLayers = &cli.IntFlag{
		Name:        "n-gpu-layers",
		Aliases:     []string{"ngl"},
		Usage:       "When compiled with GPU support, this option allows offloading some layers to the GPU for computation. Generally results in increased performance.",
		Value:       DefaultNGpuLayers,
		Destination: &Conf.NGpuLayers,
	}

	NPredict = &cli.IntFlag{
		Name:        "n-predict",
		Aliases:     []string{"n"},
		Usage:       "Set the number of tokens to predict when generating text. Adjusting this value can influence the length of the generated text.",
		Value:       defaultNPredict,
		Destination: &Conf.NPredict,
	}

	Seed = &cli.UintFlag{
		Name:        "seed",
		Aliases:     []string{"s"},
		Usage:       "Set the random number generator (RNG) seed (default: -1, -1 = random seed).",
		Value:       math.MaxUint32,
		Destination: &Conf.Seed,
	}

	Pooling = &cli.StringFlag{
		Name:        "pooling",
		Aliases:     []string{"o"},
		Usage:       "pooling type for embeddings, use model default if unspecified {none,mean,cls,last,rank}",
		Value:       "mean",
		Destination: &Conf.Pooling,
	}

	BatchSize = &cli.IntFlag{
		Name:        "batch-size",
		Aliases:     []string{"b"},
		Usage:       "logical maximum batch size",
		Value:       2048,
		Destination: &Conf.BatchSize,
	}

	UBatchSize = &cli.IntFlag{
		Name:        "ubatch-size",
		Aliases:     []string{"ub"},
		Usage:       "physical maximum batch size",
		Value:       512,
		Destination: &Conf.UBatchSize,
	}

	OutputFile = &cli.StringFlag{
		Name:        "output-file",
		Aliases:     []string{"of"},
		Usage:       "output file",
		Destination: &Conf.OutputFile,
	}

	Host = &cli.StringFlag{
		Name:        "host",
		Aliases:     []string{"ho"},
		Usage:       fmt.Sprintf("IP Address for the llama.go server (default %s)", DefaultHost),
		Value:       DefaultHost,
		EnvVars:     []string{"LLAMAGO_HOST"},
		Destination: &Conf.Host,
	}

	Origins = &cli.StringFlag{
		Name:        "origins",
		Aliases:     []string{"or"},
		Usage:       "A comma separated list of allowed origins",
		EnvVars:     []string{"LLAMAGO_ORIGINS"},
		Destination: &Conf.Origins,
	}

	Jinja = &cli.BoolFlag{
		Name:        "jinja",
		Aliases:     []string{"j"},
		Usage:       "use jinja template for chat (default: disabled)",
		Value:       false,
		Destination: &Conf.Jinja,
	}

	ChatTemplate = &cli.StringFlag{
		Name:        "chat-template",
		Aliases:     []string{"ct"},
		Usage:       "set custom jinja chat template (default: template taken from model's metadata)",
		EnvVars:     []string{"LLAMAGO_JINJA_TEMPLATE"},
		Destination: &Conf.ChatTemplate,
	}

	ChatTemplateFile = &cli.StringFlag{
		Name:        "chat-template-file",
		Aliases:     []string{"ctf"},
		Usage:       "set custom jinja chat template file (default: template taken from model's metadata)",
		EnvVars:     []string{"LLAMAGO_JINJA_TEMPLATE_FILE"},
		Destination: &Conf.ChatTemplateFile,
	}

	ChatTemplateKwargs = &cli.StringFlag{
		Name:        "chat-template-kwargs",
		Aliases:     []string{"ctk"},
		Usage:       "sets additional params for the json template parser",
		EnvVars:     []string{"LLAMAGO_CHAT_TEMPLATE_KWARGS"},
		Destination: &Conf.ChatTemplateKwargs,
	}

	NoPrune = &cli.BoolFlag{
		Name:        "noprune",
		Aliases:     []string{"np"},
		Usage:       "Do not prune model blobs on startup",
		Value:       false,
		EnvVars:     []string{"LLAMAGO_NOPRUNE"},
		Destination: &Conf.NoPrune,
	}

	AppFlags = []cli.Flag{
		LogLevel,
		Model,
		ModelDir,
		CtxSize,
		Prompt,
		NGpuLayers,
		NPredict,
		Seed,
		Pooling,
		BatchSize,
		UBatchSize,
		OutputFile,
		Host,
		Origins,
		Jinja,
		ChatTemplate,
		ChatTemplateFile,
		ChatTemplateKwargs,
		NoPrune,
	}
)

type Config struct {
	LogLevel string
	Model    string
	ModelDir string

	CtxSize            int
	Prompt             string
	NGpuLayers         int
	NPredict           int
	Seed               uint
	Pooling            string
	BatchSize          int
	UBatchSize         int
	OutputFile         string
	Host               string
	Origins            string
	Jinja              bool
	ChatTemplate       string
	ChatTemplateFile   string
	ChatTemplateKwargs string
	NoPrune            bool
}

func (c *Config) Load() error {
	log.Debug("Try to load config")
	if !c.HasModel() {
		return fmt.Errorf("No config model")
	}
	log.Debug("Model info", "model path", c.ModelPath())
	return nil
}

func (c *Config) ModelPath() string {
	if len(c.Model) <= 0 {
		return ""
	}
	if !strings.Contains(c.Model, EXT) && !strings.Contains(c.Model, EXT_BIN) {
		return ""
	}
	if common.IsFilePath(c.Model) {
		return c.Model
	}
	ret, err := filepath.Abs(filepath.Join(c.ModelDir, c.Model))
	if err != nil {
		return ""
	}
	return ret
}

func (c *Config) HasModel() bool {
	return len(c.ModelPath()) > 0
}

func (c *Config) GetModelFileInfos() []os.FileInfo {
	var firstInfo os.FileInfo
	if c.HasModel() {
		info, err := os.Stat(c.ModelPath())
		if err == nil {
			firstInfo = info
		}
	}
	hasAdd := false
	ret := []os.FileInfo{}

	if common.IsExist(c.ModelDir) {
		entries, err := os.ReadDir(c.ModelDir)
		if err != nil {
			log.Error(err.Error())
		} else {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				if filepath.Ext(entry.Name()) != EXT {
					continue
				}
				info, err := entry.Info()
				if err != nil {
					log.Error(err.Error())
					continue
				}
				ret = append(ret, info)
				if firstInfo != nil && !hasAdd {
					if os.SameFile(firstInfo, info) {
						hasAdd = true
					}
				}
			}
		}
	}

	if firstInfo != nil && !hasAdd {
		ret = append(ret, firstInfo)
	}
	return ret
}

func (c *Config) GetModelPath(model string) string {
	if len(model) <= 0 {
		return ""
	}
	if !strings.Contains(model, EXT) {
		return ""
	}
	if common.IsFilePath(model) {
		return model
	}
	ret, err := filepath.Abs(filepath.Join(c.ModelDir, model))
	if err != nil {
		return ""
	}
	return ret
}

func (c *Config) HostURL() *url.URL {
	defaultPort := DefaultPort
	chost := c.Host
	scheme, hostport, ok := strings.Cut(chost, "://")
	switch {
	case !ok:
		scheme, hostport = "http", chost
	case scheme == "http":
		defaultPort = "80"
	case scheme == "https":
		defaultPort = "443"
	}

	hostport, path, _ := strings.Cut(hostport, "/")
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		host, port = "127.0.0.1", defaultPort
		if ip := net.ParseIP(strings.Trim(hostport, "[]")); ip != nil {
			host = ip.String()
		} else if hostport != "" {
			host = hostport
		}
	}

	if n, err := strconv.ParseInt(port, 10, 32); err != nil || n > 65535 || n < 0 {
		log.Warn("invalid port, using default", "port", port, "default", defaultPort)
		port = defaultPort
	}

	return &url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(host, port),
		Path:   path,
	}
}

// AllowedOrigins returns a list of allowed origins. AllowedOrigins can be configured via the LLAMAGO_ORIGINS environment variable.
func (c *Config) AllowedOrigins() []string {
	origins := []string{}
	if len(c.Origins) > 0 {
		origins = strings.Split(c.Origins, ",")
	}

	for _, origin := range []string{"localhost", "127.0.0.1", "0.0.0.0"} {
		origins = append(origins,
			fmt.Sprintf("http://%s", origin),
			fmt.Sprintf("https://%s", origin),
			fmt.Sprintf("http://%s", net.JoinHostPort(origin, "*")),
			fmt.Sprintf("https://%s", net.JoinHostPort(origin, "*")),
		)
	}

	origins = append(origins,
		"app://*",
		"file://*",
		"tauri://*",
		"vscode-webview://*",
		"vscode-file://*",
	)

	return origins
}
