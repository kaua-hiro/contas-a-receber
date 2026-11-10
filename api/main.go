package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kaua-hiro/contas-a-receber/api/internal/titulo"
)

func main() {
	porta := env("PORT", "8080")
	origem := env("CORS_ORIGIN", "http://localhost:5173")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	titulo.NovoHandler(titulo.NovoMemoriaStore()).Registrar(mux)

	srv := &http.Server{
		Addr:              ":" + porta,
		Handler:           logar(cors(origem, mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("API rodando em http://localhost:%s", porta)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("erro ao subir servidor: %v", err)
	}
}

// cors libera o front (outra origem/porta) a chamar a API.
func cors(origem string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origem)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// logar registra método, caminho e duração de cada requisição.
func logar(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(inicio))
	})
}

func env(chave, padrao string) string {
	if v := os.Getenv(chave); v != "" {
		return v
	}
	return padrao
}
