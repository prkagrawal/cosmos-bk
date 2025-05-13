package main

import (
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/prkagrawal/cosmos-bk2/auth"
	"github.com/prkagrawal/cosmos-bk2/database"
	"github.com/prkagrawal/cosmos-bk2/graph"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/ast"

	_ "github.com/99designs/gqlgen/graphql"
)

const defaultPort = "8080"

func main() {
	// Initialize logger
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Logger()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Fatal().Err(err).Msg("Error loading .env file")
	}

	// Initialize database
	if err := database.Connect(); err != nil {
		logger.Fatal().Err(err).Msg("Database connection failed")
	}

	if err := database.Migrate(); err != nil {
		logger.Fatal().Err(err).Msg("Database migration failed")
	}

	// Create router
	router := chi.NewRouter()

	// Add middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(zerologLogger(&logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// Initialize auth service
	authSvc := auth.NewAuthService(database.DB)

	// Setup routes
	router.HandleFunc("/auth/google", auth.GoogleLoginHandler)
	router.HandleFunc("/auth/google/callback", auth.GoogleCallbackHandler(authSvc))

	// Create the main resolver, passing in dependencies
	resolver := graph.NewResolver(database.DB)

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	// Configure transports
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// Configure extensions
	if os.Getenv("ENVIRONMENT") == "development" {
		srv.Use(extension.Introspection{})
		srv.Use(extension.AutomaticPersistedQuery{
			Cache: lru.New[string](100),
		})
	}

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.FixedComplexityLimit(300))

	// Setup GraphQL routes
	router.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	router.Handle("/query", srv)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	logger.Info().Msgf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Fatal().Err(err).Msg("Server failed to start")
	}
}

// zerologLogger middleware for chi
func zerologLogger(logger *zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				// Log panics
				if rec := recover(); rec != nil {
					logger.Error().
						Str("path", r.URL.Path).
						Interface("recover_info", rec).
						Msg("Panic occurred")
					http.Error(ww, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}

				// Log the request
				logger.Info().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Int("status", ww.Status()).
					Dur("duration", time.Since(start)).
					Msg("Request handled")
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
