package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	"github.com/gatie-io/gatie-server/internal/auth"
)

type testOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

func newTestRouter(jwtManager *auth.JWTManager) *chi.Mux {
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test", "1.0.0"))

	huma.Register(api, huma.Operation{
		OperationID: "test",
		Method:      http.MethodGet,
		Path:        "/test",
		Middlewares: huma.Middlewares{NewAuthMiddleware(jwtManager)},
	}, func(ctx context.Context, input *struct{}) (*testOutput, error) {
		out := &testOutput{}
		out.Body.Message = "ok"
		return out, nil
	})

	return router
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	jwtManager := auth.NewJWTManager("secret", 15*time.Minute, 7*24*time.Hour)
	router := newTestRouter(jwtManager)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	jwtManager := auth.NewJWTManager("secret", 15*time.Minute, 7*24*time.Hour)
	router := newTestRouter(jwtManager)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("secret", 15*time.Minute, 7*24*time.Hour)
	router := newTestRouter(jwtManager)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("secret", 15*time.Minute, 7*24*time.Hour)
	router := newTestRouter(jwtManager)

	token, err := jwtManager.GenerateAccessToken("member-123", "ADMIN", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		body, _ := io.ReadAll(w.Body)
		t.Errorf("expected 200, got %d: %s", w.Code, string(body))
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("secret", -1*time.Second, 7*24*time.Hour)
	router := newTestRouter(jwtManager)

	token, _ := jwtManager.GenerateAccessToken("member-123", "ADMIN", "admin")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
