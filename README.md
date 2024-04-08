# Ollamax

Ollamax is a simple and easy to use library for making a local LLM app.  
It is based on the [Ollama](https://github.com/ollama/ollama)

## Some use case
- [Local LLM Chat Demo](demo/chat)
- [Local LLM Embedding](demo/embedding)
- [Local LLM Vision](demo/vision)

## How to use
go >= 1.22

1. Init your go module
    ```bash
    go mod init <module-name>
    ```
2. Add submodule
    ```bash
    git submodule add https://github.com/byebyebruce/ollamax.git
    ```
   
3. Init go work
    ```bash
    go work init . ./ollamax ./ollamax/ollama
    ```

4. Build library
    ```bash
    make -C ollamax
    ```
5. Write a test code
```go
import(
    "github.com/ollamax/ollamax"
)

if err := ollamax.Init(); err != nil {
    log.Fatalln(err)
}
defer ollamax.Cleanup()
llm, err := ollamax.NewWithAutoDownload("qwen:0.5b")
if err != nil {
    panic(err)
}
defer llm.Close()
llm.Chat...
```

## Where to find models
https://ollama.com/library