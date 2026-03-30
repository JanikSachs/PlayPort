package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/JanikSachs/PlayPort/internal/auth"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

func setupTestAuthHandlers(t *testing.T) (*AuthHandlers, *auth.InMemoryStateStore, *storage.InMemoryUserStore, *auth.InMemorySessionStore) {
	stateStore := auth.NewInMemoryStateStore()
	userStore := storage.NewInMemoryUserStore()
	sessionStore := auth.NewInMemorySessionStore(0)

	templates, err := template.ParseGlob("../../web/templates/*.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	ah := NewAuthHandlers(nil, nil, stateStore, userStore, sessionStore, templates, false, false)
	return ah, stateStore, userStore, sessionStore
}

func TestHandleLoginPage(t *testing.T) {
	ah, _, _, _ := setupTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	ah.HandleLoginPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Log In") {
		t.Error("Response should contain 'Log In'")
	}
	if !strings.Contains(body, "csrf_token") {
		t.Error("Response should contain CSRF token field")
	}
}

func TestHandleRegisterPage(t *testing.T) {
	ah, _, _, _ := setupTestAuthHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	w := httptest.NewRecorder()

	ah.HandleRegisterPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Register") {
		t.Error("Response should contain 'Register'")
	}
}

func TestHandleRegister_Success(t *testing.T) {
	ah, stateStore, _, sessionStore := setupTestAuthHandlers(t)

	csrfToken, _ := stateStore.Generate()

	form := url.Values{
		"csrf_token": {csrfToken},
		"username":   {"testuser"},
		"password":   {"password123"},
		"confirm":    {"password123"},
	}

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	ah.HandleRegister(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("Expected redirect to /, got %s", location)
	}

	// Should have set a session cookie
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_token" {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Should have set session_token cookie")
	}

	// Session should be valid
	userID, err := sessionStore.Get(sessionCookie.Value)
	if err != nil {
		t.Fatalf("Session should be valid: %v", err)
	}

	if userID == "" {
		t.Error("Session should contain a user ID")
	}
}

func TestHandleRegister_DuplicateUsername(t *testing.T) {
	ah, stateStore, userStore, _ := setupTestAuthHandlers(t)

	// Create existing user
	hash, _ := auth.HashPassword("password123")
	_, err := userStore.CreateWithCredentials("testuser", hash)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	csrfToken, _ := stateStore.Generate()

	form := url.Values{
		"csrf_token": {csrfToken},
		"username":   {"testuser"},
		"password":   {"password123"},
		"confirm":    {"password123"},
	}

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	ah.HandleRegister(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (re-render form), got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "already taken") {
		t.Error("Response should contain 'already taken' error")
	}
}

func TestHandleRegister_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		password    string
		confirm     string
		expectedErr string
	}{
		{
			name:        "Short username",
			username:    "ab",
			password:    "password123",
			confirm:     "password123",
			expectedErr: "between 3 and 32",
		},
		{
			name:        "Invalid username characters",
			username:    "test user!",
			password:    "password123",
			confirm:     "password123",
			expectedErr: "letters, numbers, and underscores",
		},
		{
			name:        "Short password",
			username:    "testuser",
			password:    "short",
			confirm:     "short",
			expectedErr: "at least 8 characters",
		},
		{
			name:        "Mismatched passwords",
			username:    "testuser",
			password:    "password123",
			confirm:     "password456",
			expectedErr: "do not match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ah, stateStore, _, _ := setupTestAuthHandlers(t)
			csrfToken, _ := stateStore.Generate()

			form := url.Values{
				"csrf_token": {csrfToken},
				"username":   {tt.username},
				"password":   {tt.password},
				"confirm":    {tt.confirm},
			}

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			ah.HandleRegister(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 (re-render form), got %d", w.Code)
			}

			body := w.Body.String()
			if !strings.Contains(body, tt.expectedErr) {
				t.Errorf("Response should contain '%s', got: %s", tt.expectedErr, body)
			}
		})
	}
}

func TestHandleLogin_Success(t *testing.T) {
	ah, stateStore, userStore, sessionStore := setupTestAuthHandlers(t)

	// Create a user
	hash, _ := auth.HashPassword("password123")
	_, err := userStore.CreateWithCredentials("testuser", hash)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	csrfToken, _ := stateStore.Generate()

	form := url.Values{
		"csrf_token": {csrfToken},
		"username":   {"testuser"},
		"password":   {"password123"},
	}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	ah.HandleLogin(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("Expected redirect to /, got %s", location)
	}

	// Session should be valid
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_token" {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Should have set session_token cookie")
	}

	_, err = sessionStore.Get(sessionCookie.Value)
	if err != nil {
		t.Fatalf("Session should be valid: %v", err)
	}
}

func TestHandleLogin_WrongPassword(t *testing.T) {
	ah, stateStore, userStore, _ := setupTestAuthHandlers(t)

	hash, _ := auth.HashPassword("password123")
	_, err := userStore.CreateWithCredentials("testuser", hash)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	csrfToken, _ := stateStore.Generate()

	form := url.Values{
		"csrf_token": {csrfToken},
		"username":   {"testuser"},
		"password":   {"wrongpassword"},
	}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	ah.HandleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (re-render form), got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Invalid username or password") {
		t.Error("Response should contain generic error message")
	}
}

func TestHandleLogin_NonexistentUser(t *testing.T) {
	ah, stateStore, _, _ := setupTestAuthHandlers(t)

	csrfToken, _ := stateStore.Generate()

	form := url.Values{
		"csrf_token": {csrfToken},
		"username":   {"nonexistent"},
		"password":   {"password123"},
	}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	ah.HandleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (re-render form), got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Invalid username or password") {
		t.Error("Response should contain generic error, not reveal which field is wrong")
	}
}

func TestHandleLogout(t *testing.T) {
	ah, _, _, sessionStore := setupTestAuthHandlers(t)

	// Create a session
	token, err := sessionStore.Create("user123")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	w := httptest.NewRecorder()

	ah.HandleLogout(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Expected redirect to /login, got %s", location)
	}

	// Session should be deleted
	_, err = sessionStore.Get(token)
	if err == nil {
		t.Error("Session should be deleted after logout")
	}

	// Cookie should be cleared (MaxAge=-1)
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "session_token" {
			if c.MaxAge != -1 {
				t.Errorf("session_token cookie should have MaxAge=-1, got %d", c.MaxAge)
			}
			break
		}
	}
}
