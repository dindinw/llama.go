package routes

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Qitmeer/llama.go/api"
	config2 "github.com/Qitmeer/llama.go/app/embedding/config"
	"github.com/Qitmeer/llama.go/config"
	"github.com/Qitmeer/llama.go/model"
	"github.com/Qitmeer/llama.go/version"
	"github.com/Qitmeer/llama.go/wrapper"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gin-gonic/gin"
)

const (
	InvalidModelNameErrMsg = "invalid model name"
)

func (s *API) VersionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": version.String()})
}

func (s *API) HealthHandler(c *gin.Context) {
	c.String(http.StatusOK, "Llama.go is running")
}

func (s *API) PullHandler(c *gin.Context) {
	var req api.PullRequest
	err := c.ShouldBindJSON(&req)
	switch {
	case errors.Is(err, io.EOF):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	case err != nil:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Parse model reference
	hf, err := model.ParseHuggingFaceModel(req.Model)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("failed to parse model reference: %w", err)})
		return
	}

	ch := make(chan any)
	go func() {
		defer close(ch)
		fn := func(r api.ProgressResponse) {
			ch <- r
		}

		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()

		if err := PullModel(ctx, hf, fn); err != nil {
			ch <- gin.H{"error": err.Error()}
		}
	}()

	if req.Stream != nil && !*req.Stream {
		waitForStream(c, ch)
		return
	}

	streamHandler(c, ch)

	c.Header("Content-Type", "application/json")
	var latest api.ProgressResponse
	c.JSON(http.StatusOK, latest)
}

func (s *API) PsHandler(c *gin.Context) {
	models := []api.ProcessModelResponse{}
	slices.SortStableFunc(models, func(i, j api.ProcessModelResponse) int {
		// longest duration remaining listed first
		return cmp.Compare(j.ExpiresAt.Unix(), i.ExpiresAt.Unix())
	})

	c.JSON(http.StatusOK, api.ProcessResponse{Models: models})
}

func (s *API) GenerateHandler(c *gin.Context) {
	bodyBytes, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//bodyStr := string(bodyBytes)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req api.GenerateRequest
	if err := c.ShouldBindJSON(&req); errors.Is(err, io.EOF) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, ch := wrapper.NewChan()
	if id == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "task id error"})
		return
	}
	stream := true
	if req.Stream != nil {
		stream = *req.Stream
	}
	go func() {
		m := s.cfg.GetModelPath(req.Model)
		if len(m) <= 0 {
			m = s.cfg.ModelPath()
		}
		err = s.runnerSer.Generate(id, m, req.Prompt, stream)
		if err != nil {
			log.Warn(err.Error())
			return
		}
	}()

	if !stream {
		content := ""
		for rr := range ch {
			str, ok := rr.(string)
			if !ok {
				continue
			}
			content += str
		}
		if len(content) <= 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no content"})
			return
		}
		var ret map[string]interface{}
		if err := json.Unmarshal([]byte(content), &ret); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid json"})
			return
		}
		c.JSON(http.StatusOK, ret)

		return
	}
	streamHandler(c, ch)
}

func (s *API) ChatHandler(c *gin.Context) {
	bodyBytes, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bodyStr := string(bodyBytes)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req api.ChatRequest
	if err := c.ShouldBindJSON(&req); errors.Is(err, io.EOF) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, ch := wrapper.NewChan()
	if id == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "task id error"})
		return
	}
	go func() {
		m := s.cfg.GetModelPath(req.Model)
		if len(m) <= 0 {
			m = s.cfg.ModelPath()
		}
		err = s.runnerSer.Chat(id, m, bodyStr)
		if err != nil {
			log.Warn(err.Error())
			return
		}
	}()

	if req.Stream == nil || !*req.Stream {
		content := ""
		for rr := range ch {
			str, ok := rr.(string)
			if !ok {
				continue
			}
			content += str
		}
		if len(content) <= 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no content"})
			return
		}
		var ret map[string]interface{}
		if err := json.Unmarshal([]byte(content), &ret); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid json"})
			return
		}
		c.JSON(http.StatusOK, ret)

		return
	}
	streamHandler(c, ch)
}

func (s *API) EmbedHandler(c *gin.Context) {
	checkpointStart := time.Now()
	var req api.EmbedRequest
	err := c.ShouldBindJSON(&req)
	switch {
	case errors.Is(err, io.EOF):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	case err != nil:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var input []string

	switch i := req.Input.(type) {
	case string:
		if len(i) > 0 {
			input = append(input, i)
		}
	case []any:
		for _, v := range i {
			if _, ok := v.(string); !ok {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input type"})
				return
			}
			input = append(input, v.(string))
		}
	default:
		if req.Input != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input type"})
			return
		}
	}

	checkpointLoaded := time.Now()

	if len(input) == 0 {
		c.JSON(http.StatusOK, api.EmbedResponse{Model: req.Model, Embeddings: [][]float32{}})
		return
	}

	prompts := ""
	for k, i := range input {
		if k > 0 {
			prompts += config2.Conf.EmbdSeparator
		}
		prompts += i
	}

	ret, err := wrapper.LlamaEmbedding(s.cfg, prompts, "array")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": strings.TrimSpace(err.Error())})
		return
	}
	var embeddings [][]float32
	err = json.Unmarshal([]byte(ret), &embeddings)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": strings.TrimSpace(err.Error())})
		return
	}
	if len(embeddings) != len(input) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%d != %d", len(embeddings), len(input))})
		return
	}
	resp := api.EmbedResponse{
		Model:           req.Model,
		Embeddings:      embeddings,
		TotalDuration:   time.Since(checkpointStart),
		LoadDuration:    checkpointLoaded.Sub(checkpointStart),
		PromptEvalCount: len(input),
	}
	c.JSON(http.StatusOK, resp)
}

func (s *API) EmbeddingsHandler(c *gin.Context) {
	var req api.EmbeddingRequest
	if err := c.ShouldBindJSON(&req); errors.Is(err, io.EOF) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// an empty request loads the model
	if req.Prompt == "" {
		c.JSON(http.StatusOK, api.EmbeddingResponse{Embedding: []float64{}})
		return
	}

	ret, err := wrapper.LlamaEmbedding(s.cfg, req.Prompt, "array")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": strings.TrimSpace(err.Error())})
		return
	}
	var embeddings [][]float64
	err = json.Unmarshal([]byte(ret), &embeddings)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": strings.TrimSpace(err.Error())})
		return
	}
	resp := api.EmbeddingResponse{
		Embedding: embeddings[0],
	}
	c.JSON(http.StatusOK, resp)
}

func (s *API) ListHandler(c *gin.Context) {
	models := []api.ListModelResponse{}

	infos := s.cfg.GetModelFileInfos()

	for _, info := range infos {

		models = append(models, api.ListModelResponse{
			Model:      info.Name(),
			Name:       info.Name(),
			Size:       info.Size(),
			ModifiedAt: info.ModTime(),
			Details: api.ModelDetails{
				Format: config.EXT[1:],
			},
		})
	}

	slices.SortStableFunc(models, func(i, j api.ListModelResponse) int {
		// most recently modified first
		return cmp.Compare(j.ModifiedAt.Unix(), i.ModifiedAt.Unix())
	})
	c.JSON(http.StatusOK, api.ListResponse{Models: models})
}

func (s *API) ShowHandler(c *gin.Context) {
	var req api.ShowRequest
	err := c.ShouldBindJSON(&req)
	switch {
	case errors.Is(err, io.EOF):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	case err != nil:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	showModel := s.cfg.Model
	if len(req.Model) > 0 {
		showModel = req.Model
	}
	infos := s.cfg.GetModelFileInfos()

	capabilities := []model.Capability{}
	capabilities = append(capabilities, model.CapabilityThinking)

	for _, info := range infos {
		if info.Name() != showModel {
			continue
		}
		resp := &api.ShowResponse{
			Modelfile: info.Name(),
			Details: api.ModelDetails{
				Format: config.EXT[1:],
			},
			ModifiedAt:   info.ModTime(),
			Capabilities: capabilities,
		}
		c.JSON(http.StatusOK, resp)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", showModel)})
}

// V1ModelsWebUIHandler lists models in the shape expected by llama.cpp tools/server/webui (data[] with path, status, in_cache).
func (s *API) V1ModelsWebUIHandler(c *gin.Context) {
	type statusObj struct {
		Value string `json:"value"`
	}
	type dataEntry struct {
		ID      string    `json:"id"`
		Object  string    `json:"object"`
		Created int64     `json:"created"`
		OwnedBy string    `json:"owned_by"`
		InCache bool      `json:"in_cache"`
		Path    string    `json:"path"`
		Status  statusObj `json:"status"`
	}
	activePath := s.cfg.ModelPath()
	infos := s.cfg.GetModelFileInfos()
	entries := make([]dataEntry, 0, len(infos))
	for _, info := range infos {
		path := s.cfg.GetModelPath(info.Name())
		if path == "" {
			path = filepath.Join(s.cfg.ModelDir, info.Name())
		}
		st := "unloaded"
		if activePath != "" && path == activePath {
			st = "loaded"
		}
		owned := "local"
		if hf, err := model.ParseHuggingFaceModel(info.Name()); err == nil {
			owned = hf.Namespace
		}
		entries = append(entries, dataEntry{
			ID:      info.Name(),
			Object:  "model",
			Created: info.ModTime().Unix(),
			OwnedBy: owned,
			InCache: true,
			Path:    path,
			Status:  statusObj{Value: st},
		})
	}
	slices.SortStableFunc(entries, func(i, j dataEntry) int {
		return cmp.Compare(j.Created, i.Created)
	})
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   entries,
	})
}

func (s *API) ModelsLoadStubHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *API) ModelsUnloadStubHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *API) PropsHandler(c *gin.Context) {
	status, jsonStr := wrapper.LlamaPropsHTTP()
	if status == 0 {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "llama core is not running"})
		return
	}
	if jsonStr == "" {
		c.AbortWithStatusJSON(status, gin.H{"error": "empty response from llama core"})
		return
	}
	c.Data(status, "application/json; charset=utf-8", []byte(jsonStr))
}

func (s *API) PropsChangeHandler(c *gin.Context) {
	if !wrapper.GetCommonParams().EndpointProps {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "This server does not support changing global properties. Start it with `--props`"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (s *API) SlotsHandler(c *gin.Context) {
	status, jsonStr := wrapper.LlamaSlotsHTTP()
	if status == 0 {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "llama core is not running"})
		return
	}
	if jsonStr == "" {
		c.AbortWithStatusJSON(status, gin.H{"error": "empty response from llama core"})
		return
	}
	c.Data(status, "application/json; charset=utf-8", []byte(jsonStr))
}
