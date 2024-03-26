package ollamax

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ollama/ollama/api"
	"github.com/stretchr/testify/require"
)

const (
	model = "qwen:0.5b"
)

func TestMain(m *testing.M) {

	err := Init()
	if err != nil {
		panic(err)
	}
	defer Cleanup()

	m.Run()
}

func TestCore_Predict(t *testing.T) {
	c, err := NewWithAutoDownload(model)
	defer c.Close()
	require.Nil(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	msg := []api.Message{{"user", "hello", nil}}
	for i := 0; i < 1; i++ {

		msg = append(msg, api.Message{"user", "why the sky is blue" + strconv.Itoa(i), nil})
		cc, err := c.ChatStream(ctx, msg, nil)
		require.Nil(t, err)
		require.NotNil(t, cc)
		result := ""
		for m := range cc {
			if m.Err != nil {
				fmt.Println(m.Err)
			}
			result += m.Result.Content
		}
		fmt.Println(result)

		msg = append(msg, api.Message{"assistant", result, nil})
	}
}

func TestCore_List(t *testing.T) {
	a, err := ListModel()
	require.Nil(t, err)
	require.NotNil(t, a)
	fmt.Println(a)
}

func TestCore_Pull(t *testing.T) {
	err := DeleteModel(model)
	require.Nil(t, err)
	//model = "nomic-embed-text"
	err = PullModel(context.Background(), model, func(r api.ProgressResponse) {
		fmt.Println(r.Completed, r.Total)
	})
	require.Nil(t, err)

	a, err := ListModel()
	require.Nil(t, err)
	require.NotNil(t, a)
	fmt.Println(a)
}

func TestCore_Delete(t *testing.T) {
	err := DeleteModel(model)
	require.Nil(t, err)
}

func TestCore_Embed(t *testing.T) {
	embedModel := "nomic-embed-text"

	c, err := NewWithAutoDownload(embedModel)
	defer c.Close()
	require.Nil(t, err)

	b, err := c.Embedding(context.Background(), "hello")
	require.Nil(t, err)
	require.NotNil(t, b)
	fmt.Println(b)
}
