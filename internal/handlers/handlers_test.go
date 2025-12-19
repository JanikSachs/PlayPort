package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/JanikSachs/PlayPort/internal/providers"
	"github.com/JanikSachs/PlayPort/internal/services"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

func setupTestHandlers(t *testing.T) *Handlers {
	// Create transfer service with mock provider
	transferService := services.NewTransferService()
	mockProvider := providers.NewMockProvider()
	transferService.RegisterProvider(mockProvider)
	
	// Create connection store
	connectionStore := storage.NewInMemoryConnectionStore()
	
	// Parse templates
	templates, err := template.ParseGlob("../../web/templates/*.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}
	
	return NewHandlers(transferService, templates, connectionStore, false)
}

func TestHandleHome(t *testing.T) {
	handlers := setupTestHandlers(t)
	
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	
	handlers.HandleHome(w, req)
	
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "PlayPort") {
		t.Error("Response should contain 'PlayPort'")
	}
}

func TestHandleHome_NotFound(t *testing.T) {
	handlers := setupTestHandlers(t)
	
	req := httptest.NewRequest(http.MethodGet, "/invalid", nil)
	w := httptest.NewRecorder()
	
	handlers.HandleHome(w, req)
	
	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestHandleProviders(t *testing.T) {
	handlers := setupTestHandlers(t)
	
	req := httptest.NewRequest(http.MethodGet, "/providers", nil)
	w := httptest.NewRecorder()
	
	handlers.HandleProviders(w, req)
	
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "Mock Music") {
		t.Error("Response should contain 'Mock Music' provider")
	}
}

func TestHandleTransfer(t *testing.T) {
	handlers := setupTestHandlers(t)
	
	req := httptest.NewRequest(http.MethodGet, "/transfer", nil)
	w := httptest.NewRecorder()
	
	handlers.HandleTransfer(w, req)
	
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "Transfer") {
		t.Error("Response should contain 'Transfer'")
	}
}

func TestHandleGetPlaylists(t *testing.T) {
	handlers := setupTestHandlers(t)
	
	tests := []struct {
		name           string
		provider       string
		expectedStatus int
	}{
		{
			name:           "Valid provider",
			provider:       "Mock Music",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing provider",
			provider:       "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid provider",
			provider:       "Invalid Provider",
			expectedStatus: http.StatusNotFound,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqURL := "/api/playlists"
			if tt.provider != "" {
				reqURL += "?provider=" + url.QueryEscape(tt.provider)
			}
			
			req := httptest.NewRequest(http.MethodGet, reqURL, nil)
			w := httptest.NewRecorder()
			
			handlers.HandleGetPlaylists(w, req)
			
			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			
			if tt.expectedStatus == http.StatusOK {
				body := w.Body.String()
				if !strings.Contains(body, "Summer Vibes") {
					t.Error("Response should contain playlist names")
				}
			}
		})
	}
}

func TestHandleStartTransfer(t *testing.T) {
	handlers := setupTestHandlers(t)
	
	tests := []struct {
		name           string
		method         string
		formData       url.Values
		expectedStatus int
	}{
		{
			name:   "Valid transfer request",
			method: http.MethodPost,
			formData: url.Values{
				"source_provider": []string{"Mock Music"},
				"target_provider": []string{"Mock Music"},
				"playlist_id":     []string{"mock-1"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			formData:       url.Values{},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "Missing parameters",
			method: http.MethodPost,
			formData: url.Values{
				"source_provider": []string{"Mock Music"},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/transfer/start", strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			
			handlers.HandleStartTransfer(w, req)
			
			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			
			if tt.expectedStatus == http.StatusOK {
				body := w.Body.String()
				if !strings.Contains(body, "Transfer") {
					t.Error("Response should contain transfer status")
				}
			}
		})
	}
}
