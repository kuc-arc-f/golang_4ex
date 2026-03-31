package main

import (
	"database/sql"
    "encoding/json"
    "fmt"
	"log"
    "os"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
    fmt.Println("全引数:", os.Args)
    var arg_len = len(os.Args)
    fmt.Println("arg_len=", arg_len)
    sqlite_vec.Auto()
    //db, err := sql.Open("sqlite3", ":memory:")
    db, err := sql.Open("sqlite3", "example.db")
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
        embedding FLOAT[8],
        metadata TEXT
    );`

    _, err = db.Exec(createTableSQL)
    if err != nil {
        log.Fatalf("Failed to create table: %v", err)
    }
    fmt.Println("✅ Virtual table 'vec_items' created successfully.")	

	// 3. ベクトルデータの登録 (Insert)
	// JSON 文字列としてベクトルを渡すことができます

    if arg_len >= 2 && os.Args[1] == "create"{
        insertSQL := `INSERT INTO vec_items (id, embedding, metadata) VALUES (?, ?, ?)`

        items := []struct {
            ID        int
            Embedding []float32
            Meta      string
        }{
            {1, []float32{-0.200, 0.250, 0.341, -0.211, 0.645, 0.935, -0.316, -0.924}, "item A"},
            {2, []float32{0.443, -0.501, 0.355, -0.771, 0.707, -0.708, -0.185, 0.362}, "item B"},
            {3, []float32{0.716, -0.927, 0.134, 0.052, -0.669, 0.793, -0.634, -0.162}, "item C"},
            {4, []float32{-0.710, 0.330, 0.656, 0.041, -0.990, 0.726, 0.385, -0.958}, "item D"},
        }

        for _, item := range items {
            // スライスを JSON 文字列に変換
            vecJSON, err := json.Marshal(item.Embedding)
            if err != nil {
                log.Fatalf("JSON marshal error: %v", err)
            }

            _, err = db.Exec(insertSQL, item.ID, string(vecJSON), item.Meta)
            if err != nil {
                log.Fatalf("Failed to insert item %d: %v", item.ID, err)
            }
        }
        fmt.Printf("✅ Inserted %d vectors.\n", len(items))

    }
    if arg_len >= 2 && os.Args[1] == "search" {
        // 4. ベクトル検索 (KNN Query)
        // 検索したいクエリベクトル
        queryVector := []float32{0.890, 0.544, 0.825, 0.961, 0.358, 0.0196, 0.521, 0.175}
        queryVecJSON, _ := json.Marshal(queryVector)

        // SQL クエリ: MATCH 演算子を使用して距離を計算し、ORDER BY distance でソート
        searchSQL := `
        SELECT 
            id, 
            metadata,
            distance 
        FROM vec_items 
        WHERE embedding MATCH ? 
        ORDER BY distance 
        LIMIT 2;`

        rows, err := db.Query(searchSQL, string(queryVecJSON))
        if err != nil {
            log.Fatalf("Search query error: %v", err)
        }
        defer rows.Close()

        fmt.Println("\n Search Results (Top 2):")
        fmt.Println("--------------------------")
        for rows.Next() {
            var id int
            var meta string
            var distance float64

            if err := rows.Scan(&id, &meta, &distance); err != nil {
                log.Fatalf("Scan error: %v", err)
            }
            fmt.Printf("ID: %d | Meta: %s | Distance: %.6f\n", id, meta, distance)
        }

        if err = rows.Err(); err != nil {
            log.Fatalf("Rows iteration error: %v", err)
        }
    }

}