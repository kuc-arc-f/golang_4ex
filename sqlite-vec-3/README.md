# sqlite-vec-3 

 Version: 0.9.1

 Author  :

 date    : 2026/03/31

 update :

***

Golang + sqlite-vec example

* Embedding-model : Qwen3-Embedding-0.6B-Q8_0.gguf
* llama.cpp , llama-server 
* go version go1.26.1 linux/amd64

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
go mod init example.com/sqlite-vec-3
go mod tidy

go build
```

***
* vector data add
```
./sqlite-vec-3 create
```

* search
```
./sqlite-vec-3 search hello
```

***
### blog

***

