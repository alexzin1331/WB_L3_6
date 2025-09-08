package server

import (
	"net/http"
	"strconv"
	"time"

	"L3_6/internal/storage"
	"L3_6/models"

	"github.com/gin-gonic/gin"
)

type Server struct {
	storage *storage.Storage
	router  *gin.Engine
}

func NewServer(storage *storage.Storage) *Server {
	server := &Server{storage: storage}
	server.setupRouter()
	return server
}

func (s *Server) setupRouter() {
	r := gin.Default()

	// Serve static files
	r.Static("/web", "./web")

	// API routes
	api := r.Group("/api")
	{
		api.POST("/items", s.createSale)
		api.GET("/items", s.getSales)
		api.PUT("/items/:id", s.updateSale)
		api.DELETE("/items/:id", s.deleteSale)
		api.GET("/analytics", s.getAnalytics)
		api.GET("/export", s.exportCSV)
	}

	s.router = r
}

func (s *Server) Run(port string) error {
	return s.router.Run(":" + port)
}

func (s *Server) createSale(c *gin.Context) {
	var sale models.Sale
	if err := c.ShouldBindJSON(&sale); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.storage.CreateSale(&sale); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sale)
}

func (s *Server) getSales(c *gin.Context) {
	sales, err := s.storage.GetSales()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sales)
}

func (s *Server) updateSale(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var sale models.Sale
	if err := c.ShouldBindJSON(&sale); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sale.ID = id
	if err := s.storage.UpdateSale(&sale); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sale)
}

func (s *Server) deleteSale(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := s.storage.DeleteSale(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (s *Server) getAnalytics(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from date"})
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to date"})
		return
	}

	analytics, err := s.storage.GetAnalytics(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func (s *Server) exportCSV(c *gin.Context) {
	// Implementation for CSV export
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}
