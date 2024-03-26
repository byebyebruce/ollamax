package ollamax

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"runtime"
	"sync"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/gpu"
	"github.com/ollama/ollama/llm"
	"github.com/ollama/ollama/server"
	"github.com/schollz/progressbar/v3"
)

func init() {
	err := initializeKeypair()
	if err != nil {
		log.Fatal(err)
	}
}

// Init 初始化llm环境
func Init() error {
	return llm.Init()
}

// Cleanup 清理环境
func Cleanup() {
	gpu.Cleanup()
}

// Ollamax ollama的主要结构
type Ollamax struct {
	mu    sync.Mutex
	model *llmWrapper
}

// New 创建一个新的Ollamax实例
func New(model string) (*Ollamax, error) {
	if runtime.GOOS == "linux" {
		// check compatibility to log warnings
		if _, err := gpu.CheckVRAM(); err != nil {
			slog.Info(err.Error())
		}
	}

	cm, err := load(model)
	if err != nil {
		return nil, err
	}
	c := &Ollamax{
		model: cm,
	}

	return c, nil
}

// NewWithAutoDownload 创建一个新的Ollamax实例，如果本地没有这个模型自动下载模型
func NewWithAutoDownload(model string) (*Ollamax, error) {
	has, err := HasModel(model)
	if err != nil {
		return nil, err
	}
	if !has {
		bar := progressbar.Default(100)
		err = PullModel(context.Background(), model, func(r api.ProgressResponse) {
			if r.Total == 0 {
				return
			}
			bar.Set(int(r.Completed * 100 / r.Total))
		})
		bar.Finish()
		if err != nil {
			return nil, err
		}
	}

	return New(model)
}

// Close 关闭Ollamax实例
func (c *Ollamax) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.model != nil {
		c.model.Close()
	}
}

// Reload 重新加载模型
func (c *Ollamax) Reload(model string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.model != nil && c.model.Name == model {
		return nil
	}

	_, err := server.GetModel(model)
	if err != nil {
		return err
	}

	if c.model != nil {
		c.model.Close()
	}
	c.model = nil
	lw, err := load(model)
	if err != nil {
		return err
	}
	c.model = lw
	return nil
}

type chatResponse struct {
	Result llm.PredictResult
	Err    error
}

// Chat 聊天
func (c *Ollamax) Chat(ctx context.Context, messages []api.Message) (string, error) {
	ch, err := c.ChatStream(ctx, messages)
	if err != nil {
		return "", err
	}

	fullText := ""
	for {
		select {
		case <-ctx.Done():
			return fullText, ctx.Err()
		case r, ok := <-ch:
			if !ok {
				return fullText, nil
			}
			if r.Err != nil {
				return fullText, r.Err
			}
			fullText += r.Result.Content
		}
	}
}

// ChatStream 聊天
func (c *Ollamax) ChatStream(ctx context.Context, messages []api.Message) (<-chan *chatResponse, error) {
	return c.ChatWithOption(ctx, messages, nil)
}

// ChatWithOption 聊天
func (c *Ollamax) ChatWithOption(ctx context.Context, messages []api.Message, options map[string]interface{}) (<-chan *chatResponse, error) {
	c.mu.Lock()

	if nil == c.model {
		c.mu.Unlock()
		return nil, fmt.Errorf("model is nil")
	}

	if c.model.IsEmbedding() {
		c.mu.Unlock()
		return nil, fmt.Errorf("model is not an embedding model")
	}

	opts, err := modelOptions(c.model.Model, options)
	if err != nil {
		c.mu.Unlock()
		return nil, err
	}

	encode := func(s string) ([]int, error) {
		return c.model.Encode(ctx, s)
	}
	prompt, err := server.ChatPrompt(c.model.Template, messages, opts.NumCtx, encode)
	if err != nil {
		c.mu.Unlock()
		return nil, err
	}

	ch := make(chan *chatResponse)
	//checkpointStart := time.Now()
	format := ""
	if options != nil {
		if v, ok := options["format"]; ok {
			format = v.(string)
		}
	}
	// only support json format
	if len(format) > 0 && format != "json" {
		return nil, fmt.Errorf("format %s not supported", format)
	}

	go func() {
		defer c.mu.Unlock()
		defer close(ch)

		fn := func(r llm.PredictResult) {
			select {
			case <-ctx.Done():
				return
			case ch <- &chatResponse{r, nil}:
			}
		}

		// Start prediction
		predictReq := llm.PredictOpts{
			Prompt: prompt,
			Format: format,
			//Images:  images,
			Options: opts,
		}
		if err := c.model.Predict(ctx, predictReq, fn); err != nil {
			select {
			case <-ctx.Done():
			case ch <- &chatResponse{Err: err}:
			}
			return
		}
	}()
	return ch, nil
}

// Embedding 嵌入
func (c *Ollamax) Embedding(ctx context.Context, prompt string) ([]float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if nil == c.model {
		return nil, fmt.Errorf("model is nil")
	}
	if !c.model.IsEmbedding() {
		return nil, fmt.Errorf("model is not an embedding model")
	}

	embedding, err := c.model.Embedding(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return embedding, nil
}
