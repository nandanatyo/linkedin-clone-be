package routes

import (
	"github.com/gin-gonic/gin"
	"linked-clone/internal/middleware"
)

func UserRoutes(rg *gin.RouterGroup, deps *Dependencies) {
	authMiddleware := middleware.AuthMiddleware(deps.JWTService, deps.Logger)

	users := rg.Group("/users")
	{
		// Public endpoints
		users.GET("/search", deps.UserHandler.SearchUsers)
		users.GET("/:id", deps.UserHandler.GetUserByID)

		// Profile endpoints
		users.GET("/profile", authMiddleware, deps.UserHandler.GetProfile)
		users.PUT("/profile", authMiddleware, deps.UserHandler.UpdateProfile)
		users.POST("/profile/picture",
			authMiddleware,
			middleware.FileUploadMiddleware(5<<20, []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}),
			deps.UserHandler.UploadProfilePicture,
		)

		// Connection endpoints - NEW
		connections := users.Group("/connections", authMiddleware)
		{
			// Send connection request
			connections.POST("/request", deps.ConnectionHandler.SendConnectionRequest)

			// Manage connection requests
			connections.POST("/:id/accept", deps.ConnectionHandler.AcceptConnectionRequest)
			connections.POST("/:id/reject", deps.ConnectionHandler.RejectConnectionRequest)
			connections.DELETE("/:id", deps.ConnectionHandler.RemoveConnection)

			// Get connections and requests
			connections.GET("", deps.ConnectionHandler.GetUserConnections)
			connections.GET("/requests", deps.ConnectionHandler.GetConnectionRequests)
			connections.GET("/sent", deps.ConnectionHandler.GetSentRequests)

			// Check connection status with specific user
			connections.GET("/status/:userId", deps.ConnectionHandler.GetConnectionStatus)
			connections.GET("/mutual/:userId", deps.ConnectionHandler.GetMutualConnections)

			// Block/unblock users
			connections.POST("/block/:userId", deps.ConnectionHandler.BlockUser)
			connections.DELETE("/block/:userId", deps.ConnectionHandler.UnblockUser)
		}
	}
}
