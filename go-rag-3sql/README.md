# go-rag-3sql

 Version: 0.9.1

 Author  :

 date    : 2026/06/02

 update :

***

Golang Windows + SQLite example

* embedding: Gemini-embedding-001
* model: Gemma-4-E2B
* llama.cpp , llama-server 

***

### setup

* llama-server start
* port 8090: gemma-4-E2B

```
#gemma-4-E2B

/usr/local/llama-b8642/llama-server -m /var/lm_data/unsloth/gemma-4-E2B-it-Q4_K_S.gguf \
 --chat-template-kwargs '{"enable_thinking": false}' --port 8090 

```

***
### .env

```
GEMINI_API_KEY=
```

***
### related

https://huggingface.co/unsloth/gemma-4-E2B-it-GGUF

***
### build
```
go mod init example.com/go-rag-3sql
go mod tidy

go build
```

***
* vector data add

```
sqlite3 ./example.db < table.sql
```
***
* create
```
go-rag-3sql.exe create
```

* search
```
go-rag-3sql.exe search hello
```

***
### blog

***

