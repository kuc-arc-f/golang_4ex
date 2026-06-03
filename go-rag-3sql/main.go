package main

import (
    "bytes"
    //"context"
    "database/sql"
    "encoding/json"
    //"flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "example.com/go-rag-3sql/handler"

    "github.com/joho/godotenv"
    _ "modernc.org/sqlite"
)

type ReadParam struct {
    Content  string    `json:"content"`
    Name     string    `json:"name"`
}

// RequestPayload represents the JSON payload for the Ollama API
type RequestPayload struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}
// ResponsePayload represents the JSON response from the Ollama API
type ResponsePayload struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}
// リクエストの構造体
type EmbeddingRequest struct {
    Model   string  `json:"model"`
    Content EmbeddingContent `json:"content"`
}

type EmbeddingContent struct {
    Parts []EmbeddingPart `json:"parts"`
}

type EmbeddingPart struct {
    Text string `json:"text"`
}
// レスポンスの構造体
type EmbeddingResponse struct {
    Embedding Embedding `json:"embedding"`
}
type Embedding struct {
    Values []float32 `json:"values"`
}

// Collection APIレスポンス用構造体
type Collection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// 検索用リクエスト構造体
type QueryRequest struct {
	QueryEmbeddings [][]float32 `json:"query_embeddings"`
	NResults        int         `json:"n_results"`
	Include         []string    `json:"include,omitempty"`
}

// 検索結果レスポンス用構造体
type QueryResponse struct {
	IDs       [][]string                 `json:"ids"`
	Distances [][]float64                `json:"distances"`
	Metadatas [][]map[string]interface{} `json:"metadatas"`
	Documents [][]string                 `json:"documents"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}


type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

func send_chat(query string) string{
    var input = "日本語で、回答して欲しい。\n 要約して欲しい。\n" + query
    fmt.Printf("input: \n%v\n\n", input)

    history := []Message{
        {
            Role:    "system",
            Content: "You are a helpful assistant. 日本語で答えてください。",
        },
    }
    history = append(history, Message{
        Role:    "user",
        Content: input,
    })    
    var serverURL   = "http://localhost:8090/v1/chat/completions"
    var model       = "local-model"
    var temperature = 0.7

    reqBody := ChatRequest{
        Model:       model,
        Messages:    history,
        Temperature: temperature,
    }

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("JSONマーシャルエラー:", err)
		return ""
	}
	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("リクエスト送信エラー:", err)
		return ""
	}
	defer resp.Body.Close()

    // レスポンスボディの読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("レスポンス読み取りエラー: %v\n", err)
		return ""
	}
	
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
        fmt.Errorf("JSONデコードエラー: %w", err)
		return "" 
	}
	if len(chatResp.Choices) == 0 {
        fmt.Errorf("レスポンスにChoicesがありません")
		return "" 
	}

    var outStr string = chatResp.Choices[0].Message.Content;
    //fmt.Printf("\n outStr %s\n\n", outStr)
    return outStr;
}

const DB_NAME = "example.db"

/**
*
* @param
*
* @return
*/
func searchQuery(query string, apiKey string){
    // テキストの埋め込みを取得
    text := query
    embeddings, err := handler.GetEmbeddings(text, apiKey)
    if err != nil {
        fmt.Printf("エラーが発生しました: %v\n", err)
        return
    }           
    // 結果の出力
    fmt.Println("\n取得したベクトルデータ:")
    fmt.Printf("次元数: %d\n", len(embeddings))  

    var outText string = handler.CheckSimalirity(query, embeddings, DB_NAME)
    //var input string = ""
    //input = "日本語で、回答して欲しい。\n" + outText
    var out_str = send_chat(outText)

    fmt.Printf("AI:\n%s", out_str)
    return
}
/**
*
* @param
*
* @return
*/
func main() {
    err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}    
	// 環境変数からAPIキーを取得
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("エラー: 環境変数 GEMINI_API_KEY が設定されていません")
        return
	}
    db, err := sql.Open("sqlite", DB_NAME)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()    

    var query = "二十四節季"
    fmt.Println("query:", query)
    fmt.Println("全引数:", os.Args)
    if len(os.Args) > 1 {
        fmt.Println("最初の引数:", os.Args[1])
        var argment = os.Args[1]
        if argment == "search" {
            var query = os.Args[2]
            fmt.Println("query:", query)
            searchQuery(query ,apiKey)
        }
        if argment == "create" {
            fmt.Println("create-start:")
            handler.CreateVector(apiKey, DB_NAME)
        }
    } else {
        fmt.Println("none , arg ")
    }
}
