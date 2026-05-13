package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/aigate/model"
)

func APIKeyAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": map[string]interface{}{
				"message": "missing authorization header",
				"type":    "invalid_request_error",
			}})
			return
		}

		rawKey := strings.TrimPrefix(authHeader, "Bearer ")
		if rawKey == authHeader || !strings.HasPrefix(rawKey, "sk-") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": map[string]interface{}{
				"message": "invalid api key format",
				"type":    "invalid_request_error",
			}})
			return
		}

		h := sha256.Sum256([]byte(rawKey))
		keyHash := hex.EncodeToString(h[:])

		var apiKey model.APIKey
		if err := db.Where("key_hash = ?", keyHash).First(&apiKey).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": map[string]interface{}{
				"message": "invalid api key",
				"type":    "invalid_request_error",
			}})
			return
		}

		if apiKey.Status != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": map[string]interface{}{
				"message": "api key has been disabled",
				"type":    "invalid_request_error",
			}})
			return
		}

		if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": map[string]interface{}{
				"message": "api key has expired",
				"type":    "invalid_request_error",
			}})
			return
		}

		c.Set("api_key_id", apiKey.ID)
		c.Set("user_id", apiKey.UserID)
		c.Set("tenant_id", apiKey.TenantID)
		c.Set("api_key", &apiKey)

		go db.Model(&model.APIKey{}).Where("id = ?", apiKey.ID).Update("last_used_at", time.Now())

		c.Next()
	}
}
