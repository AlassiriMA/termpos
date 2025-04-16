package main

import (
        "database/sql"
        "fmt"
        "log"

        _ "github.com/mattn/go-sqlite3"
)

func main() {
        db, err := sql.Open("sqlite3", "./pos.db")
        if err != nil {
                log.Fatal(err)
        }
        defer db.Close()

        rows, err := db.Query("PRAGMA table_info(sales);")
        if err != nil {
                log.Fatal(err)
        }
        defer rows.Close()

        fmt.Println("Sales Table Schema:")
        fmt.Println("------------------------")
        for rows.Next() {
                var cid, notnull, pk int
                var name, ctype string
                var dflt_value interface{}
                
                if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
                        log.Fatal(err)
                }
                
                fmt.Printf("%d: %s (%s)\n", cid, name, ctype)
        }
        
        if err := rows.Err(); err != nil {
                log.Fatal(err)
        }

        // Check if we have any customers
        rows, err = db.Query("SELECT COUNT(*) FROM customers;")
        if err != nil {
                log.Fatal(err)
        }
        defer rows.Close()

        var customerCount int
        if rows.Next() {
                if err := rows.Scan(&customerCount); err != nil {
                        log.Fatal(err)
                }
        }
        
        fmt.Printf("\nFound %d customer(s) in the database\n", customerCount)
        
        // If no customers, let's add a test customer
        if customerCount == 0 {
                _, err := db.Exec(`
                        INSERT INTO customers (
                                name, email, phone, join_date, loyalty_points, loyalty_tier, 
                                created_at, updated_at
                        ) VALUES (
                                'Test Customer', 'test@example.com', '555-1234', 
                                CURRENT_TIMESTAMP, 100, 'Bronze', 
                                CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
                        )
                `)
                if err != nil {
                        log.Fatal(err)
                }
                fmt.Println("Added a test customer to the database")
        }
}