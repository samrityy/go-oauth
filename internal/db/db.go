package db
import (
    "context"
    "database/sql"
    "os"
    "fmt"
    _ "github.com/jackc/pgx/v5/stdlib" 
)

func DB() (*sql.DB, error) {
    dsn := os.Getenv("POSTGRES_DSN") // e.g. "postgres://user:pass@localhost:5432/dbname?sslmode=disable"
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, err
    }
    fmt.Println("Connecting to database with DSN:", dsn)

    // Optional: set max open and idle connections
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(0)

    // Ping to verify connection
    ctx := context.Background()
    if err := db.PingContext(ctx); err != nil {
        return nil, err
    }
    return db, nil
}
