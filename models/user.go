package models

// User represents a row in the users table.
// Password field intentionally has no json tag so it is omitted when marshaling to JSON.
type User struct {
	ID       int64  `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `db:"password"`
	Email    string `json:"email" db:"email"`
	FullName string `json:"full_name" db:"full_name"`
	Role     string `json:"role" db:"role"`
}

