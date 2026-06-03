package handler

import (
    "bytes"
    //"context"
    "database/sql"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "math"
    "net/http"
    "os"
    "path/filepath"
    "sort"

    "github.com/google/uuid"
    "github.com/tmc/langchaingo/textsplitter"
    _ "modernc.org/sqlite"
)
const DATA_DIR = "./data"
const CHUNK_SIZE_MAX = 500

type ReadParam struct {
    Content  string    `json:"content"`
    Name     string    `json:"name"`
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
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func GetEmbeddings(inputText string, apiKey string) ([]float32, error) {
    //var ret []float32 
	// リクエストボディの作成
	reqBody := EmbedContentRequest{
		Model: "models/gemini-embedding-001",
		Content: Content{
			Parts: []Part{
				{Text: inputText},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("JSONマーシャルエラー: %v", err)
	}

	// HTTPリクエストの作成
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("リクエスト作成エラー: %v", err)
	}

	// ヘッダーの設定
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

    
	// リクエスト送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("HTTPリクエストエラー: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスボディの読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("レスポンス読み取りエラー: %v", err)
	}

	// ステータスコードチェック
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("APIエラー (status=%d): %s", resp.StatusCode, string(body))
	}

	// JSONパース
	var embedResp EmbedContentResponse
	if err := json.Unmarshal(body, &embedResp); err != nil {
		log.Fatalf("JSONアンマーシャルエラー: %v", err)
	}

	// 結果の出力
	fmt.Printf("✅ 埋め込みベクトル取得成功\n")
	fmt.Printf("次元数: %d\n", len(embedResp.Embedding.Values))
	
	// ベクトルの先頭5要素を表示（確認用）
	displayCount := 5
	if len(embedResp.Embedding.Values) < displayCount {
		displayCount = len(embedResp.Embedding.Values)
	}
	fmt.Printf("値 (先頭%d件): %v\n", displayCount, embedResp.Embedding.Values[:displayCount])
	// 最初の要素の embedding ベクトルを返す
	// 配列が複数ある場合はインデックスを調整してください
	return embedResp.Embedding.Values , nil
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

// --- リクエスト用構造体 ---

type EmbedContentRequest struct {
	Model   string  `json:"model"`
	Content Content `json:"content"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

// --- レスポンス用構造体 ---
type EmbedContentResponse struct {
	Embedding Embedding `json:"embedding"`
}

/**
*
* @param
*
* @return
 */
func add_vec (vec[]float32 , content string , uuid string , db_path string )bool {
    var ret bool = false

    var s_buff  = "["
    for i, row := range vec {
        if i > 0 {
            s_buff += ","
        }
        s_buff += fmt.Sprintf("%f", row)
    }
    s_buff += "]"
    //db-add
    db, err := sql.Open("sqlite", db_path)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()    

    result, err := db.Exec(
        "INSERT INTO document(id, content, embeddings) VALUES(?, ?, ?)",
        uuid,
        content,
        s_buff,
    )
    if err != nil {
        log.Fatal(err)
        return ret
    }

    id, _ := result.LastInsertId()
    fmt.Println("Insert ID =", id)     
    //fmt.Printf("s_buff=%v\n", s_buff)

    ret = true
    return ret
}
/**
*
* @param
*
* @return
 */
func CreateVector(apiKey string , db_path string) {
    fmt.Printf("#CreateVector-start\n")

    var fileItems []ReadParam = readTextData()
    fmt.Printf("len=%v\n", len(fileItems))

    for i, fileRow := range fileItems {
        fmt.Printf("i=%d, name=%v\n", i, fileRow.Name)
        fmt.Printf("con.len=%d\n", len(fileRow.Content))
        newID := uuid.New().String()
        fmt.Printf("newID=%v\n", newID)

        // 関数呼び出し
        embeddings, err := GetEmbeddings(fileRow.Content , apiKey)
        if err != nil {
            fmt.Printf("エラーが発生しました: %v\n", err)
            return
        }           
        // 結果の出力
        fmt.Println("\n取得したベクトルデータ:")
        fmt.Printf("次元数: %d\n", len(embeddings)) 
        
        var ret_add = add_vec(embeddings , fileRow.Content , newID, db_path)
        if !ret_add {
            fmt.Println("error, insert add_vec \n")
            return
        }
    }
    fmt.Println("データ挿入完了")
}

/**
*
* @param
*
* @return
*/
func cosineSimilarity(a []float32, b []float32) (float64, error) {
    if len(a) != len(b) {
        return 0, fmt.Errorf("vectors must have the same length")
    }

    var dotProduct, aMagnitude, bMagnitude float64
    for i := 0; i < len(a); i++ {
        dotProduct += float64(a[i] * b[i])
        aMagnitude += float64(a[i] * a[i])
        bMagnitude += float64(b[i] * b[i])
    }

    if aMagnitude == 0 || bMagnitude == 0 {
        return 0, nil
    }

    return dotProduct / (math.Sqrt(aMagnitude) * math.Sqrt(bMagnitude)), nil
}
/**
*
* @param
*
* @return
*/
func convertFloat32(value []byte) []float32 {
    var float64s []float64
    if err := json.Unmarshal(value, &float64s); err != nil {
        panic(err)
    }        
    float32s := make([]float32, len(float64s))
    for i, v := range float64s {
        float32s[i] = float32(v)
    }        
    //fmt.Printf("float32s.len= %v\n", len(float32s))
    return float32s
}
/**
*
* @param
*
* @return
*/
func CheckSimalirity(query string , vec []float32, db_path string) string {

    type OutEmbed struct {
        Embed  []byte    `json:"embeddings"`
        Content string   `json:"content"`
        Id string   `json:"id"`
    }
    type ScoreEmbed struct {
        Embed  []byte    `json:"embeddings"`
        Content string   `json:"content"`
        Id string   `json:"id"`
        Score float64 `json:"score"`
    }

    db, err := sql.Open("sqlite", db_path)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()    

	select_sql := fmt.Sprintf(`SELECT id, content, embeddings FROM document`)
    fmt.Printf("sql=%s\n", select_sql)
	rows, err := db.Query(select_sql)
	if err != nil {
        log.Fatal(err)
	}
	defer rows.Close()

    var matches string = ""
	var outData []OutEmbed
	var scoreData []ScoreEmbed
	for rows.Next() {
		var row OutEmbed
        var scoreRow ScoreEmbed
		err := rows.Scan(&row.Id, &row.Content, &row.Embed)
		if err != nil {
          log.Fatal(err)
		}
        //fmt.Printf("embed=%v\n", row.Embed)
        float32s := convertFloat32(row.Embed)
        similarity, _ := cosineSimilarity(vec, float32s)
        fmt.Printf("sim= %v id=%s\n", similarity, row.Id)
        if(similarity > 0.6) {
            scoreRow.Content = row.Content
            scoreRow.Id = row.Id
            scoreRow.Score = similarity
            scoreData = append(scoreData, scoreRow)
        }
		outData = append(outData, row)
	}
    sort.Slice(scoreData, func(i, j int) bool {
		// 大きい値でソート（降順）したい場合は `>` を使います
		return scoreData[i].Score > scoreData[j].Score
	})    
    fmt.Println("\n=== ソート後（Scoreの大きい順） ===")
    var topNum int =1
    var addCount int = 0
    for _, v := range scoreData {
        //fmt.Printf("ID: %d, Score: %.1f\n", v.Id, v.Score)
        fmt.Printf("ID: %d, Score: %f\n", v.Id, v.Score)
        if addCount < topNum {
            matches += v.Content + "\n"
        }
        addCount += 1
    }
    //fmt.Printf("matches=%v\n", matches)
    var outText string = ""
    if (len(matches) > 0){
        outText = `context:` + matches + "\n"
        outText += `user query:` + query + "\n"
    }else{
        outText =`user query:` + query + "\n"
    }
    fmt.Printf("outText=%v\n", outText)
    return outText
}