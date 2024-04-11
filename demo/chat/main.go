package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/byebyebruce/ollamax"
	"github.com/fatih/color"
	"github.com/ollama/ollama/api"
)

var (
	model = flag.String("model", "qwen:0.5b", "model to use")
)

func main() {
	flag.Parse()

	slog.SetLogLoggerLevel(slog.LevelWarn)
	if err := ollamax.Init(); err != nil {
		log.Fatalln(err)
	}
	defer ollamax.Cleanup()

	o, err := ollamax.NewWithAutoDownload(*model)
	if err != nil {
		panic(err)
	}
	defer o.Close()

	go func() {
		stdIn := bufio.NewReader(os.Stdin)
		history := []api.Message{}
		for {
			for len(history) > 40 {
				history = history[2:]
			}
			fmt.Println("You:")
			i, err := stdIn.ReadString('\n')
			if err != nil {
				return
			}
			if len(strings.TrimSpace(i)) == 0 {
				continue
			}
			outChan, err := o.ChatStream(context.Background(), append(history, api.Message{"user", i, nil}))
			if err != nil {
				fmt.Println(err)
				return
			}

			full := ""
		LOOP:
			for m := range outChan {
				if m.Err != nil {
					fmt.Println(m.Err)
					break LOOP
				}
				fmt.Print(color.GreenString(m.Result.Content))
				full += m.Result.Content
			}
			fmt.Println()
			history = append(history, api.Message{"user", i, nil}, api.Message{"assistant", full, nil})
		}
	}()

	// 创建一个通道来接收操作系统信号
	sigChan := make(chan os.Signal, 1)
	// 通知信号处理程序捕获 SIGINT（Ctrl+C）
	signal.Notify(sigChan, syscall.SIGINT)
	<-sigChan // 阻塞直到收到 SIGINT
	fmt.Println("捕获到 Ctrl+C，准备退出...")
}
