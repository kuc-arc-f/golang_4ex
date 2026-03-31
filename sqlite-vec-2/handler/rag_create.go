package handler

import (
    "database/sql"
    "bytes"
    //"context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"

    //"github.com/joho/godotenv"
    "github.com/tmc/langchaingo/textsplitter"
	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

const DATA_DIR = "./data"
const CHUNK_SIZE_MAX = 1000

type ReadParam struct {
    Content  string    `json:"content"`
    Name     string    `json:"name"`
}

// EmbedRequest: llama-server に送信するリクエスト構造体
type EmbedRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// EmbedResponse: llama-server から返ってくるレスポンス構造体
// 実際のレスポンス形式は llama.cpp のバージョンにより若干異なる場合がありますが、
// 標準的な OpenAI 互換フォーマットに基づいています。
type EmbedResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func GetEmbeddings(serverURL, model, inputText string) ([]float64, error) {
	// 1. リクエストボディの作成
	reqBody := EmbedRequest{
		Input: inputText,
		Model: model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("JSON マーシャルエラー: %w", err)
	}

	// 2. HTTP リクエストの構築
	req, err := http.NewRequest("POST", serverURL+"/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成エラー: %w", err)
	}

	// ヘッダー設定
	req.Header.Set("Content-Type", "application/json")

	// 3. HTTP クライアントの実行
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("リクエスト実行エラー: %w", err)
	}
	defer resp.Body.Close()

	// ステータスコードの確認
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API エラー (Status: %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// 4. レスポンスの解析
	var embedResp EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("レスポンス解析エラー: %w", err)
	}

	// データが存在するか確認
	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("埋込データが返されませんでした")
	}

	// 最初の要素の embedding ベクトルを返す
	// 配列が複数ある場合はインデックスを調整してください
	return embedResp.Data[0].Embedding, nil
}

/**
*
* @param
*
* @return
*/
func readTextData() []ReadParam{
    fileItem := []ReadParam{}

    entries, err := os.ReadDir(DATA_DIR)
    if err != nil {
        fmt.Println("フォルダ読み込みエラー:", err)
        return nil
    }
        // textsplitter Setting
    splitter := textsplitter.NewRecursiveCharacter(
        textsplitter.WithChunkSize(CHUNK_SIZE_MAX),
        textsplitter.WithChunkOverlap(10),
    )        

    var row ReadParam
    for _, entry := range entries {
//        if entry.IsDir() ||  (filepath.Ext(entry.Name()) != ".txt")= {
        if entry.IsDir() {
            continue
        }
        if (filepath.Ext(entry.Name()) == ".txt" || filepath.Ext(entry.Name()) == ".md") {
            path := filepath.Join(DATA_DIR, entry.Name())
                    row.Name = entry.Name()

            data, err := os.ReadFile(path)
            if err != nil {
                fmt.Println("ファイル読み込みエラー:", err)
                continue
            }
            row.Content = string(data)
            // chunks add
            chunks, err := splitter.SplitText(row.Content)
            if err != nil {
                log.Fatal(err)
            }
            for i, chunk := range chunks {
                fmt.Printf("Chunk %d:\n%s\n---\n", i+1, chunk)
                row.Content = chunk
                fileItem = append(fileItem, row)
            }    
        }
        //fmt.Printf("=== %s ===\n%s\n\n", entry.Name(), string(data))
    }
    return fileItem
}


/**
*
* @param
*
* @return
 */
func CreateVector(db_path string) {
    sqlite_vec.Auto()
    fmt.Printf("#CreateVector-start\n")
    db, err := sql.Open("sqlite3", db_path)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

	// 設定値
	serverURL := "http://localhost:8080"
	modelName := "embedding-model"
	//inputText := "これはテスト文章です"

    var fileItems []ReadParam = readTextData()
    fmt.Printf("len=%v\n", len(fileItems))

    for i, fileRow := range fileItems {
        fmt.Printf("i=%d, name=%v\n", i, fileRow.Name)
        fmt.Printf("con.len=%d\n", len(fileRow.Content))
        // 関数呼び出し
        embeddings, err := GetEmbeddings(serverURL, modelName, fileRow.Content)
        if err != nil {
            fmt.Printf("エラーが発生しました: %v\n", err)
            return
        } 
        // 結果の出力
        fmt.Println("\n取得したベクトルデータ:")
        fmt.Printf("次元数: %d\n", len(embeddings))       
        // 結果の表示
        fmt.Printf("テキスト: %s\n", fileRow.Content)

        tx, _ := db.Begin()
        result, _ := tx.Exec(
            `INSERT INTO documents (title, content) VALUES (?, ?)`,
            "タイトル",
            "内容",
        )
        lastID, _ := result.LastInsertId()
        tx.Commit()

        fmt.Printf("lastID=%d\n", lastID)
        insertSQL := `INSERT INTO vec_items (id, embedding, metadata) VALUES (?, ?, ?)`
        vecJSON, err := json.Marshal(embeddings)
        if err != nil {
            log.Fatalf("JSON marshal error: %v", err)
        }        
        _, err = db.Exec(insertSQL, lastID, string(vecJSON), fileRow.Content)
        if err != nil {
            log.Fatalf("Failed to insert item %d: %v", lastID, err)
        }        
        //fmt.Printf("埋め込みベクトルの次元数: %d\n", len(result.Embedding.Values))
        //fmt.Printf("最初の5要素: %v\n", result.Embedding.Values[:5])
        //var embed_value = result.Embedding.Values
        //newID := uuid.New().String()

    }
    fmt.Println("データ挿入完了")
}
