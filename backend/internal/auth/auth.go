package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"getreleased/internal/database"
)

type Service struct {
	db        *database.DB
	jwtSecret []byte
	ttl       time.Duration
}

func NewService(db *database.DB, jwtSecret string, ttl time.Duration) (*Service, error) {
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET required")
	}
	return &Service{db: db, jwtSecret: []byte(jwtSecret), ttl: ttl}, nil
}

// EnsureAdminSeed 若 users 表为空，用环境变量密码初始化 admin 账号
func (s *Service) EnsureAdminSeed(ctx context.Context, username, password string) error {
	count, err := s.db.CountUsers(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if password == "" {
		return errors.New("ADMIN_PASSWORD required for initial admin seed (no users in DB)")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	return s.db.CreateUser(ctx, username, hash, "admin")
}

// HashPassword 用 bcrypt 哈希明文密码，供 handler 创建/重置密码时调用
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func (s *Service) VerifyPassword(ctx context.Context, username, password string) bool {
	user, err := s.db.GetUserByUsername(ctx, username)
	if err != nil || user == nil {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}

func (s *Service) IssueToken(username string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(s.ttl)
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
		Subject:   username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

func (s *Service) JWTSecret() []byte { return s.jwtSecret }

func ParseToken(tokenString string, jwtSecret []byte) (string, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}
