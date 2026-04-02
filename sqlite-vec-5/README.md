# sqlite-vec-5

 Version: 0.9.1

 Author  :

 date    : 2026/04/02

 update :

***

Golang + sqlite-vec example

* GithubCopilot CLI SDK
* Embedding-model : Qwen3-Embedding-0.6B-Q8_0.gguf
* llama.cpp , llama-server 
* go version go1.26.1 linux/amd64

***
### related

https://github.com/github/copilot-sdk

https://github.com/github/copilot-sdk/tree/main/go

***
### setup


* llama-server start
* port 8080: Qwen3-Embedding-0.6B

```
#Qwen3-Embedding-0.6B
/home/user123/llama-server -m /var/lm_data/Qwen3-Embedding-0.6B-Q8_0.gguf --embedding  -c 1024 --port 8080

```

***
* env value
```
export DATABASE_URL=./example.db
```

***
### related

https://huggingface.co/Qwen/Qwen3-Embedding-0.6B-GGUF

***
### build
```
go mod init example.com/sqlite-vec-5
go mod tidy

go build
```

***
* vector data add
```
./sqlite-vec-5 create
```

* search
```
./sqlite-vec-5 search hello
```

***
### blog

***

