package task

import (
	"errors"
	"net/http"
	"time"

	"TaskScheduler/internal/entities"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// generateID is a simple ID generator for this example
func generateID() string {
	return time.Now().Format("20060102150405")
}

func (h *Handler) CreateTask(c *gin.Context) {
	var t entities.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t.ID = generateID()
	t.CreatedAt = time.Now()
	t.Status = "pending"

	if err := h.repo.Create(&t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, t)
}

func (h *Handler) GetTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	task, err := h.repo.Get(id)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *Handler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var updateReq struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.repo.Update(id, func(t *entities.Task) {
		if updateReq.Title != nil {
			t.Title = *updateReq.Title
		}
		if updateReq.Description != nil {
			t.Description = *updateReq.Description
		}
		if updateReq.Status != nil {
			t.Status = *updateReq.Status
		}
	})

	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", h.Health)
	tasksGroup := router.Group("/tasks")
	{
		tasksGroup.POST("", h.CreateTask)
		tasksGroup.GET("/:id", h.GetTask)
		tasksGroup.PUT("/:id", h.UpdateTask)
	}
}
