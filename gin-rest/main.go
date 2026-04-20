/*
 * Project: ${projectName}
 * Template: Gin REST API – SunGo Project Manager
 *
 * Features:
 *   - Gin router with versioned API (/api/v1/...)
 *   - CORS middleware (configurable origins)
 *   - Request logger middleware
 *   - /health endpoint
 *   - Example CRUD routes for "items"
 *   - Graceful shutdown on SIGTERM / SIGINT
 */

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// ── Models ────────────────────────────────────────────────────────────────────

type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// In-memory store (replace with your DB layer)
var (
	items  = []Item{{ID: 1, Name: "example", Value: "hello from ${projectName}"}}
	nextID = 2
)

// ── Handlers ──────────────────────────────────────────────────────────────────

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"project": "${projectName}",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func listItems(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}

func getItem(c *gin.Context) {
	id := c.Param("id")
	for _, item := range items {
		if id == string(rune('0'+item.ID)) {
			c.JSON(http.StatusOK, item)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
}

func createItem(c *gin.Context) {
	var input struct {
		Name  string `json:"name"  binding:"required"`
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item := Item{ID: nextID, Name: input.Name, Value: input.Value}
	nextID++
	items = append(items, item)
	c.JSON(http.StatusCreated, item)
}

func deleteItem(c *gin.Context) {
	id := c.Param("id")
	for i, item := range items {
		if id == string(rune('0'+item.ID)) {
			items = append(items[:i], items[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"deleted": id})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
}

// ── Router ────────────────────────────────────────────────────────────────────

func setupRouter() *gin.Engine {
	r := gin.New()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	r.GET("/health", healthHandler)

	// API v1
	v1 := r.Group("/api/v1")
	{
		v1.GET("/items",      listItems)
		v1.GET("/items/:id",  getItem)
		v1.POST("/items",     createItem)
		v1.DELETE("/items/:id", deleteItem)
	}

	return r
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := setupRouter()
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("[${projectName}] listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped.")
}
