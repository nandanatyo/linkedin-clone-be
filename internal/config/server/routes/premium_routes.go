package routes

import (
	"linked-clone/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PremiumRoutes(rg *gin.RouterGroup, deps *Dependencies) {
	authMiddleware := middleware.AuthMiddleware(deps.JWTService, deps.Logger)
	premiumMiddleware := middleware.PremiumMiddleware(deps.UserRepository, deps.Logger)

	premium := rg.Group("/premium", authMiddleware)
	{

		premium.GET("/analytics", premiumMiddleware, func(c *gin.Context) {

			c.JSON(http.StatusOK, gin.H{
				"message": "Premium analytics data",
				"feature": "premium_only",
			})
		})

		premium.GET("/advanced-search", premiumMiddleware, func(c *gin.Context) {

			c.JSON(http.StatusOK, gin.H{
				"message": "Premium advanced search",
				"feature": "premium_only",
			})
		})
	}
}
