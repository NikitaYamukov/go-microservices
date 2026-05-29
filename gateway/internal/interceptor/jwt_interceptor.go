package interceptor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JWTClaims - структура для JWT claims.
type JWTClaims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

type contextKey string

const (
	UserIDKey           contextKey = "user_id"
	AuthorizationHeader            = "authorization"
	BearerPrefix                   = "Bearer "
)

// JWTInterceptor - интерсептор для локальной валидации JWT-токенов.
type JWTInterceptor struct {
	jwtSecret []byte
}

var publicMethods = map[string]bool{
	"/gateway.Gateway/Register":      true,
	"/gateway.Gateway/Login":         true,
	"/gateway.Gateway/Refresh":       true,
	"/gateway.Gateway/Logout":        true,
	"/gateway.Gateway/ValidateToken": true,
}

// NewJWTInterceptor создает новый экземпляр JWTInterceptor.
func NewJWTInterceptor(jwtSecret string) *JWTInterceptor {
	return &JWTInterceptor{jwtSecret: []byte(jwtSecret)}
}

// UnaryInterceptor - интерсептор для unary вызовов.
func (i *JWTInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		userID, err := i.extractUserIDFromJWT(ctx)
		if err != nil {
			return nil, err
		}

		ctxWithUserID := context.WithValue(ctx, UserIDKey, userID)
		return handler(ctxWithUserID, req)
	}
}

// StreamInterceptor - интерсептор для stream вызовов.
func (i *JWTInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if publicMethods[info.FullMethod] {
			return handler(srv, stream)
		}

		userID, err := i.extractUserIDFromJWT(stream.Context())
		if err != nil {
			return err
		}

		ctxWithUserID := context.WithValue(stream.Context(), UserIDKey, userID)
		wrappedStream := &wrappedServerStream{
			ServerStream: stream,
			ctx:          ctxWithUserID,
		}

		return handler(srv, wrappedStream)
	}
}

// extractUserIDFromJWT извлекает user_id из JWT-токена в метаданных.
func (i *JWTInterceptor) extractUserIDFromJWT(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	authHeaders := md.Get(AuthorizationHeader)
	if len(authHeaders) == 0 {
		return 0, status.Error(codes.Unauthenticated, "authorization header is not provided")
	}

	token := authHeaders[0]
	if !strings.HasPrefix(token, BearerPrefix) {
		return 0, status.Error(codes.Unauthenticated, "invalid token format")
	}

	accessToken := strings.TrimPrefix(token, BearerPrefix)
	if accessToken == "" {
		return 0, status.Error(codes.Unauthenticated, "access token is empty")
	}

	userID, err := i.parseAndValidateJWT(accessToken)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return userID, nil
}

// parseAndValidateJWT парсит и валидирует JWT-токен.
func (i *JWTInterceptor) parseAndValidateJWT(tokenString string) (uint64, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return i.jwtSecret, nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return 0, fmt.Errorf("token is not valid")
	}

	if userID, err := claimUserID(claims["user_id"]); err == nil && userID != 0 {
		return userID, nil
	}

	if userID, err := claimUserID(claims["sub"]); err == nil && userID != 0 {
		return userID, nil
	}

	return 0, fmt.Errorf("user_id not found in token")
}

func claimUserID(value interface{}) (uint64, error) {
	switch typed := value.(type) {
	case float64:
		if typed <= 0 || typed != float64(uint64(typed)) {
			return 0, fmt.Errorf("invalid user id")
		}
		return uint64(typed), nil
	case string:
		return strconv.ParseUint(typed, 10, 64)
	default:
		return 0, fmt.Errorf("invalid user id")
	}
}

// ValidateTokenWithoutAuthService валидирует токен без обращения к auth service.
func (i *JWTInterceptor) ValidateTokenWithoutAuthService(tokenString string) (uint64, error) {
	return i.parseAndValidateJWT(tokenString)
}

// GetUserIDFromContext извлекает user_id из контекста.
func GetUserIDFromContext(ctx context.Context) (uint64, error) {
	userID, ok := ctx.Value(UserIDKey).(uint64)
	if !ok {
		return 0, fmt.Errorf("user_id not found in context")
	}

	return userID, nil
}

// wrappedServerStream - обертка для ServerStream с кастомным контекстом.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
