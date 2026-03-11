package httpserver

import (
	"event-tracking-service/config"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	engine     *gin.Engine
	httpServer *http.Server
	config     *config.Config
}

func New(cfg *config.Config) *Server {
	engine := gin.New()

	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// Sentry middleware for panic recovery and performance monitoring
	if cfg.Sentry.Enable {
		engine.Use(sentrygin.New(sentrygin.Options{
			Repanic: true,
		}))
	}

	engine.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	return &Server{
		engine: engine,
		httpServer: &http.Server{
			Addr:    ":" + cfg.Server.Port,
			Handler: engine,
		},
		config: cfg,
	}
}

func (s *Server) Engine() *gin.Engine {
	return s.engine
}

func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
