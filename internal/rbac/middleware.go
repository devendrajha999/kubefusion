package rbac

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kubefusion/kubefusion/internal/auth"
)

func Authn(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		claims, err := auth.Parse(secret, strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func Require(roles ...string) gin.HandlerFunc {
	allow := map[string]struct{}{}
	for _, r := range roles {
		allow[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		r, _ := role.(string)
		if _, ok := allow[r]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
