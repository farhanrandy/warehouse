package config

import (
	"database/sql"
	"fmt"
	"os"

	// Third-party packages
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// OpenDB opens a connection pool to PostgreSQL using environment variables.
func OpenDB() (*sql.DB, error) {
    _ = godotenv.Load()

    host := getenv("DB_HOST", "localhost")        
    port := getenv("DB_PORT", "5432")          
    user := getenv("DB_USER", "postgres")       
    pass := getenv("DB_PASSWORD", "")         
    name := getenv("DB_NAME", "warehouse")      
    sslm := getenv("DB_SSLMODE", "disable")      

    // Build the DSN (Data Source Name) string for lib/pq.
    // lib/pq supports both URL and keyword formats; here we use keyword format.
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        host, port, user, pass, name, sslm,
    )

    // Open a database handle using the pq driver name.
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }

    // Optionally, ping the database to verify the connection parameters.
    if err := db.Ping(); err != nil {
        _ = db.Close()
        return nil, err
    }

    // Return the ready-to-use connection pool.
    return db, nil
}

// getenv reads an environment variable and falls back to a default value if missing.
func getenv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}
