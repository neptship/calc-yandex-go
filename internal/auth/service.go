package auth

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists     = errors.New("user already exists")
	ErrInvalidLogin   = errors.New("invalid login or password")
	ErrInternalServer = errors.New("internal server error")
)

var globalJWTSecret []byte

func init() {
	globalJWTSecret = make([]byte, 32)
	_, err := rand.Read(globalJWTSecret)
	if err != nil {
		log.Fatalf("Failed to generate JWT secret: %v", err)
	}
	log.Println("JWT secret generated successfully")
}

type UserClaims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

type Service struct {
	db         *sql.DB
	jwtSecret  []byte
	jwtExpires time.Duration
}

func NewService(db *sql.DB) (*Service, error) {
	return &Service{
		db:         db,
		jwtSecret:  globalJWTSecret,
		jwtExpires: 24 * time.Hour,
	}, nil
}

func (s *Service) Register(login, password string) error {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)", login).Scan(&exists)
	if err != nil {
		return ErrInternalServer
	}
	if exists {
		return ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ErrInternalServer
	}

	_, err = s.db.Exec("INSERT INTO users (login, password_hash) VALUES (?, ?)",
		login, string(hashedPassword))
	if err != nil {
		return ErrInternalServer
	}

	return nil
}

func (s *Service) Login(login, password string) (string, error) {
	var id int
	var hashedPassword string
	err := s.db.QueryRow("SELECT id, password_hash FROM users WHERE login = ?", login).Scan(&id, &hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrInvalidLogin
		}
		return "", ErrInternalServer
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return "", ErrInvalidLogin
	}

	claims := UserClaims{
		UserID: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.jwtExpires).Unix(),
			Subject:   login,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", ErrInternalServer
	}

	return tokenString, nil
}

func (s *Service) ValidateToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
