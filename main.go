package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rhruban/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set.")
	}
	secret := os.Getenv("SECRET")
	if secret == "" {
		log.Fatal("SECRET must be set.")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error opening database: %w", err)
	}
	dbQueries := database.New(dbConn)

	apicfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		secret:         secret,
	}

	mux := http.NewServeMux()
	fsHandler := apicfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("POST /api/users", apicfg.handlerUsersCreate)
	mux.HandleFunc("POST /api/login", apicfg.handlerLogin)

	mux.HandleFunc("POST /api/chirps", apicfg.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", apicfg.handlerChirpsGetAll)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apicfg.handlerChirpsGetSingle)

	mux.HandleFunc("GET /admin/metrics", apicfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apicfg.handlerReset)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
