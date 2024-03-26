.PHONY: init
init: ollama/llm/llama.cpp/build ollama/server/export.go
	@echo init

ollama/llm/llama.cpp/build: ollama/go.mod
	cd ollama && rm -rf go.sum && go mod tidy && go generate ./...

ollama/go.mod:
	git submodule update --init --depth 1 ollama

ollama/server/export.go:
	@echo write export.go
	@printf "package server\n\ntype RegistryOptions = registryOptions\n" > ollama/server/export.go

