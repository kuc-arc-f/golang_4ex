# sqlite-vec-2 

 Version: 0.9.1

 Author  :

 date    : 2026/03/31

 update :

***

Golang + sqlite-vec example

* modl: Qwen3.5-2B-Q4_K_S.gguf
* Embedding-model : Qwen3-Embedding-0.6B-Q8_0.gguf
* llama.cpp , llama-server 
* go version go1.26.1 linux/amd64

***

### setup


* llama-server start
* port 8080: Qwen3-Embedding-0.6B
* port 8090: Qwen3.5-2B

```
#Qwen3-Embedding-0.6B
/home/user123/llama-server -m /var/lm_data/Qwen3-Embedding-0.6B-Q8_0.gguf --embedding  -c 1024 --port 8080

#Qwen3.5-2B
/home/user123/llama-server -m /var/lm_data/unsloth/Qwen3.5-2B-GGUF/Qwen3.5-2B-Q4_K_S.gguf \
 --chat-template-kwargs '{"enable_thinking": false}' --port 8090 

```

***
### related

https://huggingface.co/unsloth/Qwen3.5-2B-GGUF

https://huggingface.co/Qwen/Qwen3-Embedding-0.6B-GGUF

***
### build
```
go mod init example.com/sqlite-vec-2
go mod tidy

go build
```

***
* vector data add
```
./sqlite-vec-2 create
```

* search
```
./sqlite-vec-2 search hello
```

***
### blog

https://zenn.dev/knaka0209/scraps/f096d17e22be23

***

