package repositories

import (
    "context"
    "database/sql"
    "fmt"
    "warehouse/models"
)

type UserRepo struct { DB *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{DB: db} }

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
    const q = `SELECT id, username, password, email, full_name, role FROM users WHERE username = $1`
    var u models.User
    if err := r.DB.QueryRowContext(ctx, q, username).Scan(&u.ID, &u.Username, &u.Password, &u.Email, &u.FullName, &u.Role); err != nil {
        if err == sql.ErrNoRows { return nil, nil }
        return nil, fmt.Errorf("get user by username: %w", err)
    }
    return &u, nil
}
