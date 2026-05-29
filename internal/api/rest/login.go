package rest

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubefusion/kubefusion/internal/auth"
)

func (s *Server) login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role := "viewer"
	if req.Username == "admin" {
		role = "admin"
	}
	if req.Username == "operator" {
		role = "operator"
	}
	tok, err := auth.Issue(s.JWTSecret, req.Username, role, 8*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token issue failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tok, "role": role})
}
