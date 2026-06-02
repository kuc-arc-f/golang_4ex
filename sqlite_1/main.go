package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func initDb(db *sql.DB) bool {
    var ret bool = false
    _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        age INTEGER
    )
    `)
    if err != nil {
        log.Fatal(err)
        return ret
    }
    ret = true
    return ret
}

func add_db(db *sql.DB) bool {
    var ret bool = false
    result, err := db.Exec(
        "INSERT INTO users(name, age) VALUES(?, ?)",
        "Tanaka",
        30,
    )
    if err != nil {
        log.Fatal(err)
        return ret
    }

    id, _ := result.LastInsertId()
    fmt.Println("Insert ID =", id)  
    ret = true
    return ret
}

func delete_db(db *sql.DB , id int) bool {
    var ret bool = false
    _, err := db.Exec(
        "DELETE FROM users WHERE id=?",
        id,
    )
    if err != nil {
        log.Fatal(err)
        return ret
    }
    ret = true
    return ret
}

func select_db(db *sql.DB) bool {
    var ret bool = false    
    rows, err := db.Query(
        "SELECT id, name, age FROM users",
    )
    if err != nil {
        log.Fatal(err)
        return ret
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        var name string
        var age int

        err := rows.Scan(&id, &name, &age)
        if err != nil {
            log.Fatal(err)
            return ret
        }

        fmt.Println(id, name, age)
    }    
    ret = true
    return ret
}

func main() {
    db, err := sql.Open("sqlite", "sample.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    var ret = initDb(db)
    if !ret {
        return
    }
    ret = add_db(db)
    if !ret {
        return
    }
    ret = select_db(db)
    if !ret {
        return
    }
    ret =  delete_db(db, 1)
    if !ret {
        return
    }

    fmt.Println("SQLite Open OK")
}
