package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lancer/log/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type Manager struct {
	secret []byte
	ttl    time.Duration
}

func NewManager(secret string, ttlHours int) *Manager {
	ttl := time.Duration(ttlHours) * time.Hour
	if ttl <= 0 {
		ttl = 72 * time.Hour
	}
	return &Manager{secret: []byte(secret), ttl: ttl}
}

func (m *Manager) Issue(u model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": u.ID.String(),
		"un":  u.Username,
		"exp": time.Now().Add(m.ttl).Unix(),
		"iat": time.Now().Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(m.secret)
}

var ErrInvalidToken = errors.New("invalid token")

func (m *Manager) Parse(tokenStr string) (uuid.UUID, string, error) {
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil || !tok.Valid {
		return uuid.Nil, "", ErrInvalidToken
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, "", ErrInvalidToken
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, "", ErrInvalidToken
	}
	uid, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}
	un, _ := claims["un"].(string)
	return uid, un, nil
}

// Middleware enforces a valid Bearer token; sets uid in context.
func (m *Manager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing auth"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			c.AbortWithStatusJSON(401, gin.H{"error": "bad auth header"})
			return
		}
		uid, un, err := m.Parse(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}
		c.Set("uid", uid)
		c.Set("username", un)
		c.Next()
	}
}

// HashPassword / Check for bcrypt.
func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// AuthHeader helper for the SPA to read token.
func BearerAuthHeader(token string) string {
	return "Bearer " + token
}

// SetCookieCookie for optional cookie-based auth on admin static.
func (m *Manager) SetAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "xu_admin",
		Value:     token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(m.ttl.Seconds()),
	})
}