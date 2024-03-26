package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/byebyebruce/ollamax"
)

var (
	model = flag.String("model", "nomic-embed-text", "model to use")
)

func main() {
	flag.Parse()

	if err := ollamax.Init(); err != nil {
		log.Fatalln(err)
	}
	defer ollamax.Cleanup()

	o, err := ollamax.NewWithAutoDownload(*model)
	if err != nil {
		panic(err)
	}
	defer o.Close()

	vec, err := o.Embedding(context.Background(), "hello")
	if err != nil {
		panic(err)
	}
	fmt.Println(vec)
}
