package main

import (
    "bytes"
	"database/sql"
    "encoding/json"
    "fmt"
    "io"
	"log"
    "os"
    "net/http"
    "example.com/sqlite-vec-2/handler"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

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
const (
	serverURL   = "http://localhost:8090/v1/chat/completions"
	model       = "local-model"
	temperature = 0.7
)
// レスポンス構造体
type CompletionResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
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

func main() {
    fmt.Println("全引数:", os.Args)
    var arg_len = len(os.Args)
    fmt.Println("arg_len=", arg_len)
    sqlite_vec.Auto()
    var db_path = "example.db";
    db, err := sql.Open("sqlite3", db_path)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    var vecVersion string
    err = db.QueryRow("select vec_version()").Scan(&vecVersion)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("vec_version=%s\n",vecVersion)


    // 拡張機能が正しく読み込まれているか確認するために、バージョン等をチェックしても良いですが、
    // bindings/go を import することで自動的に load_vec0 が実行される仕組みになっています。

    // 2. vec0 仮想テーブルの作成
    // float[8] のベクトル列 'embedding' を持つテーブルを作成
    createTableSQL := `
    CREATE VIRTUAL TABLE IF NOT EXISTS vec_items USING vec0(
        id INTEGER PRIMARY KEY,
        embedding FLOAT[1024],
        metadata TEXT
    );
    CREATE TABLE IF NOT EXISTS documents 
    ( 
      id INTEGER PRIMARY KEY AUTOINCREMENT, 
      title TEXT NOT NULL, 
      content TEXT NOT NULL, 
      source TEXT 
    );
    `

    _, err = db.Exec(createTableSQL)
    if err != nil {
        log.Fatalf("Failed to create table: %v", err)
    }
    fmt.Println("✅ Virtual table 'vec_items' created successfully.")	

	// 3. ベクトルデータの登録 (Insert)
    if arg_len >= 2 && os.Args[1] == "create"{
        handler.CreateVector(db_path)
    }
    if arg_len >= 3 && os.Args[1] == "search" {
        var query = os.Args[2]
        // 設定値
        serverURL := "http://localhost:8080"
        modelName := "embedding-model"      
        // 関数呼び出し
        embeddings, err := handler.GetEmbeddings(serverURL, modelName, query)
        if err != nil {
            fmt.Printf("エラーが発生しました: %v\n", err)
            return
        }           
        // 結果の出力
        fmt.Println("\n取得したベクトルデータ:")
        fmt.Printf("次元数: %d\n", len(embeddings))       

        // 4. ベクトル検索 (KNN Query)
        queryVecJSON, _ := json.Marshal(embeddings)

        // SQL クエリ: MATCH 演算子を使用して距離を計算し、ORDER BY distance でソート
        searchSQL := `
        SELECT 
            id, 
            metadata,
            distance 
        FROM vec_items 
        WHERE embedding MATCH ? 
        ORDER BY distance 
        LIMIT 1;`

        rows, err := db.Query(searchSQL, string(queryVecJSON))
        if err != nil {
            log.Fatalf("Search query error: %v", err)
        }
        defer rows.Close()

        fmt.Println("\n Search Results (Top N):")
        fmt.Println("--------------------------")
        var outStr string = "";
        var matches string = "";
        for rows.Next() {
            var id int
            var meta string
            var distance float64

            if err := rows.Scan(&id, &meta, &distance); err != nil {
                log.Fatalf("Scan error: %v", err)
            }
            if(distance < 1.0) {
                matches = meta
            }
            //fmt.Printf("ID: %d | Meta: %s | Distance: %.6f\n", id, meta, distance)
            fmt.Printf("ID: %d | Distance: %.6f\n", id, distance)
        }
        //fmt.Printf("Meta: %s\n\n", outStr)
        if err = rows.Err(); err != nil {
            log.Fatalf("Rows iteration error: %v", err)
        }
        if (len(matches) > 0){
            outStr = `context:` + matches + "\n"
            outStr += `user query:` + query + "\n"
        }else{
            outStr =`user query:` + query + "\n"
        }            
        var resp = send_chat(outStr)
        fmt.Printf("result: \n\n%s\n\n", resp)
    }

}