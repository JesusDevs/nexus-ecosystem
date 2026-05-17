package vec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Embedder genera embeddings para texto usando un modelo local.
type Embedder interface {
	Embed(text string) ([]float32, error)
	EmbedBatch(texts []string) ([][]float32, error)
	ModelName() string
	Dimension() int
}

// ── Ollama Embedder ─────────────────────────────────────────────────

type OllamaEmbedder struct {
	baseURL    string
	model      string
	dimension  int
	httpClient *http.Client
}

type OllamaConfig struct {
	BaseURL string // default: http://localhost:11434
	Model   string // default: bge-m3
}

type ollamaEmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func NewOllamaEmbedder(cfg OllamaConfig) (*OllamaEmbedder, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:11434"
	}
	if cfg.Model == "" {
		cfg.Model = "bge-m3"
	}

	e := &OllamaEmbedder{
		baseURL:   strings.TrimRight(cfg.BaseURL, "/"),
		model:     cfg.Model,
		dimension: 1024, // bge-m3 uses 1024 dimensions
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Verificar que Ollama está corriendo y tiene el modelo
	if err := e.healthCheck(); err != nil {
		return nil, fmt.Errorf("ollama no disponible: %w\n\n"+
			"  Instala Ollama: curl -fsSL https://ollama.com/install.sh | sh\n"+
			"  Descarga el modelo: ollama pull %s", err, cfg.Model)
	}

	return e, nil
}

func (e *OllamaEmbedder) healthCheck() error {
	resp, err := e.httpClient.Get(e.baseURL + "/api/tags")
	if err != nil {
		return fmt.Errorf("no se pudo conectar a Ollama en %s: %w", e.baseURL, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tags ollamaTagsResponse
	if err := json.Unmarshal(body, &tags); err != nil {
		return fmt.Errorf("respuesta inválida de Ollama: %w", err)
	}

	modelFound := false
	for _, m := range tags.Models {
		if strings.Contains(m.Name, e.model) || strings.Contains(m.Name, strings.ReplaceAll(e.model, ":", "/")) {
			modelFound = true
			break
		}
	}

	if !modelFound {
		return fmt.Errorf("modelo %s no encontrado. Ejecuta: ollama pull %s", e.model, e.model)
	}

	return nil
}

func (e *OllamaEmbedder) Embed(text string) ([]float32, error) {
	reqBody := ollamaEmbedRequest{
		Model:  e.model,
		Prompt: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := e.httpClient.Post(
		e.baseURL+"/api/embeddings",
		"application/json",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("error llamando a Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama devolvió %d: %s", resp.StatusCode, string(body))
	}

	var embResp ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("error decodificando respuesta: %w", err)
	}

	// Convertir float64 → float32
	vec := make([]float32, len(embResp.Embedding))
	for i, v := range embResp.Embedding {
		vec[i] = float32(v)
	}

	return vec, nil
}

func (e *OllamaEmbedder) EmbedBatch(texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		vec, err := e.Embed(text)
		if err != nil {
			return nil, fmt.Errorf("error en texto %d: %w", i, err)
		}
		results[i] = vec
	}
	return results, nil
}

func (e *OllamaEmbedder) ModelName() string { return e.model }
func (e *OllamaEmbedder) Dimension() int    { return e.dimension }

// ── Mock Embedder (para testing sin Ollama) ─────────────────────────

type MockEmbedder struct {
	dimension int
	rngState  uint64
}

func NewMockEmbedder(dim int) *MockEmbedder {
	return &MockEmbedder{dimension: dim, rngState: 42}
}

func (m *MockEmbedder) Embed(text string) ([]float32, error) {
	// Genera un embedding determinista basado en el texto (hash simple)
	vec := make([]float32, m.dimension)
	for i := range vec {
		h := uint64(i*31 + len(text))
		for _, c := range text {
			h = h*17 + uint64(c)
		}
		vec[i] = float32(h%1000)/1000.0 - 0.5
	}
	// Normalizar
	var norm float64
	for _, v := range vec {
		norm += float64(v * v)
	}
	norm = float64(float32(norm))
	if norm > 0 {
		scale := float32(1.0 / float32(norm))
		for i := range vec {
			vec[i] *= scale
		}
	}
	return vec, nil
}

func (m *MockEmbedder) EmbedBatch(texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i, t := range texts {
		v, _ := m.Embed(t)
		result[i] = v
	}
	return result, nil
}

func (m *MockEmbedder) ModelName() string { return "mock-embedder" }
func (m *MockEmbedder) Dimension() int    { return m.dimension }

// ── Factory ─────────────────────────────────────────────────────────

// NewEmbedderFromStore creates an embedder using configuration from the Store.
// Environment variables act as overrides: EMBEDDER_MOCK, OLLAMA_HOST, OLLAMA_EMBED_MODEL.
func NewEmbedderFromStore(store *Store) (Embedder, error) {
	if os.Getenv("EMBEDDER_MOCK") == "true" {
		dims := 1024
		if d := store.GetConfig("embed.dims"); d != "" {
			fmt.Sscanf(d, "%d", &dims)
		}
		return NewMockEmbedder(dims), nil
	}

	baseURL := os.Getenv("OLLAMA_HOST")
	if baseURL == "" {
		baseURL = store.GetConfig("ollama.host")
	}

	model := os.Getenv("OLLAMA_EMBED_MODEL")
	if model == "" {
		model = store.GetConfig("embed.model")
	}

	return NewOllamaEmbedder(OllamaConfig{
		BaseURL: baseURL,
		Model:   model,
	})
}

// NewEmbedder creates an embedder from environment variables (legacy fallback).
func NewEmbedder() (Embedder, error) {
	if os.Getenv("EMBEDDER_MOCK") == "true" {
		return NewMockEmbedder(1024), nil
	}

	baseURL := os.Getenv("OLLAMA_HOST")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	model := os.Getenv("OLLAMA_EMBED_MODEL")
	if model == "" {
		model = "bge-m3"
	}

	return NewOllamaEmbedder(OllamaConfig{
		BaseURL: baseURL,
		Model:   model,
	})
}
