// backend/auth/middleware.go
package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireAuth(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("session")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		userID, err := GetSessionUserID(db, cookie)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session invalid or expired"})
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
