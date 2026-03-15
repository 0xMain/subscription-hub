package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/0xMain/subscription-hub/internal/domain"
	"github.com/0xMain/subscription-hub/internal/pkg/strutil"

	"github.com/0xMain/subscription-hub/internal/repository/dao"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type (
	SignUpParams struct {
		Email     string
		FirstName string
		LastName  string
		Password  string
	}

	SignInResult struct {
		AccessToken string
		User        *domain.User
	}
)

type authnStore interface {
	ByEmail(ctx context.Context, email string) (*domain.UserWithPassword, error)
	Create(ctx context.Context, user *dao.User) (*domain.User, error)
}

type AuthnService struct {
	store     authnStore
	jwtSecret string
}

func NewAuthnService(store authnStore, jwtSecret string) *AuthnService {
	return &AuthnService{
		store:     store,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthnService) SignUp(ctx context.Context, p SignUpParams) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка при хешировании: %w", err)
	}

	user, err := s.store.Create(ctx,
		&dao.User{
			Email:        strutil.NormalizeEmail(p.Email),
			FirstName:    strutil.Capitalize(p.FirstName),
			LastName:     strutil.Capitalize(p.LastName),
			PasswordHash: string(hashedPassword),
		})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthnService) SignIn(ctx context.Context, email, password string) (*SignInResult, error) {
	auth, err := s.store.ByEmail(ctx, strutil.NormalizeEmail(email))
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(auth.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	tokenString, err := s.generateToken(auth.User.ID)
	if err != nil {
		return nil, err
	}

	return &SignInResult{AccessToken: tokenString, User: auth.User}, nil
}

func (s *AuthnService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, s.parseToken)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, domain.ErrInvalidToken
}

func (s *AuthnService) generateToken(userID int64) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": strconv.FormatInt(userID, 10),
		"iat": now.Unix(),
		"exp": now.Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("ошибка при подписи токена: %w", err)
	}

	return t, nil
}

func (s *AuthnService) parseToken(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("неожиданный метод при подписи токена: %v", token.Header["alg"])
	}

	return []byte(s.jwtSecret), nil
}
