package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/yaninyzwitty/cqrs-eccomerce-service/graph"
	"github.com/yaninyzwitty/cqrs-eccomerce-service/pkg"
)

func main() {
	var cfg pkg.Config
	file, err := os.Open("config.yaml")
	if err != nil {
		slog.Error("failed to open config.yaml", "error", err)
		os.Exit(1)
	}
	defer file.Close()
	if err := cfg.LoadFile(file); err != nil {
		slog.Error("failed to load config.yaml", "error", err)
		os.Exit(1)
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)

	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GraphQLServer.Port),
		Handler: mux,
	}

	stopCH := make(chan os.Signal, 1)
	signal.Notify(stopCH, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	go func() {
		slog.Info("SERVER starting", "port", cfg.GraphQLServer.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	<-stopCH
	slog.Info("shutting down the server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown server", "error", err)
		os.Exit(1)
	} else {
		slog.Info("server stopped gracefully")
	}
}
