package main

import (
	"comment-tree/internal/comment/repo/postgres"
	"comment-tree/internal/comment/rest"
	"comment-tree/internal/comment/service"
	"comment-tree/internal/config"
	"comment-tree/internal/validator"
	"comment-tree/pkg/db"
	"context"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Initialize logger
	zlog.Init()

	// Context
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Initialize config
	cfg := config.MustLoad()

	// Connect to DB
	DB, err := db.OpenDB(cfg.DB)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to DB")
	}

	// Initialize validator
	v := validator.New()

	// Initialize comment repository
	commentRepo := postgres.New(DB)

	// Initialize comment service
	commentService := service.New(commentRepo)

	// Initialize comment repository
	commentHandler := rest.NewCommentHandler(commentService, v)

	// Initialize Gin engine
	engine := ginext.New("")
	engine.Use(ginext.Logger())
	engine.Use(ginext.Recovery())
	engine.Use(CORSMiddleware())

	apiGroup := engine.Group("/api")
	apiGroup.POST("/comments", commentHandler.CreateComment)
	apiGroup.GET("/comments/:id", commentHandler.GetCommentTree)
	apiGroup.GET("/comments", commentHandler.GetComments)
	apiGroup.DELETE("/comments/:id", commentHandler.DeleteComment)

	// Initialize and start http server
	server := &http.Server{
		Addr:    cfg.Server.HTTPPort,
		Handler: engine,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("failed to listen start http server")
		}
	}()

	<-ctx.Done()

	// Graceful shutdown
	withTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.Shutdown(withTimeout); err != nil {
		zlog.Logger.Error().Err(err).Msg("server shutdown failed")
	}

	if err := DB.Master.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close master database")
	}
}

func CORSMiddleware() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
