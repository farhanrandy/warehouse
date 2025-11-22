package config

import (
	// Import the standard library packages used for DB access and env handling
	"database/sql" // database/sql provides the generic DB API
	"fmt"          // fmt is used to format strings (like the DSN)
	"os"           // os lets us read environment variables

	// Third-party packages
	"github.com/joho/godotenv" // godotenv loads variables from a .env file into environment
	_ "github.com/lib/pq"      // lib/pq is the PostgreSQL driver (blank-import to register it)
)

// OpenDB opens a connection pool to PostgreSQL using environment variables.
// It returns a *sql.DB that you should keep and reuse (it's a pooled handle).
func OpenDB() (*sql.DB, error) {
    // Attempt to load variables from a local .env file if present.
    // If the file doesn't exist, Load() returns an error we can safely ignore.
    _ = godotenv.Load()

    // Read individual environment variables for connection parameters.
    host := getenv("DB_HOST", "localhost")        // DB host, default to localhost
    port := getenv("DB_PORT", "5432")            // DB port, default to 5432
    user := getenv("DB_USER", "postgres")        // DB user, default to postgres
    pass := getenv("DB_PASSWORD", "")            // DB password, default empty
    name := getenv("DB_NAME", "warehouse")        // DB name, default to warehouse
    sslm := getenv("DB_SSLMODE", "disable")       // SSL mode, default disable for local dev

    // Build the DSN (Data Source Name) string for lib/pq.
    // lib/pq supports both URL and keyword formats; here we use keyword format.
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        host, port, user, pass, name, sslm,
    )

    // Open a database handle using the pq driver name.
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err // Return error if driver open fails (e.g., bad driver name)
    }

    // Optionally, ping the database to verify the connection parameters.
    if err := db.Ping(); err != nil {
        // If Ping fails, close the handle to avoid leaking resources and return the error.
        _ = db.Close()
        return nil, err
    }

    // Return the ready-to-use connection pool.
    return db, nil
}

// getenv reads an environment variable and falls back to a default value if missing.
func getenv(key, def string) string {
    // Look up the value from the current environment.
    if v := os.Getenv(key); v != "" {
        return v // Return the value if it is set and non-empty
    }
    return def // Otherwise, return the provided default
}
