package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/byebyebruce/ollamax"
	"github.com/fatih/color"
	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(chatCMD(), listCMD(), pullCMD())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func chatCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "chat",
		Short: "chat",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Println("need input model(eg: qwen:0.5b, qwen, gemma, llama3, phi3)")
			}
			model := args[0]
			if err := ollamax.Init(); err != nil {
				log.Fatalln(err)
			}
			defer ollamax.Cleanup()

			o, err := ollamax.NewWithAutoDownload(model)
			if err != nil {
				panic(err)
			}
			defer o.Close()

			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				stdIn := bufio.NewReader(cmd.InOrStdin())
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
					outChan, err := o.ChatStream(ctx, append(history, api.Message{"user", i, nil}))
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

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT)
			<-sigChan
			cancel()
			return nil
		},
	}
}
