package routes

import (
	"github.com/gin-gonic/gin"
	"linked-clone/internal/middleware"
)

func UserRoutes(rg *gin.RouterGroup, deps *Dependencies) {
	authMiddleware := middleware.AuthMiddleware(deps.JWTService, deps.Logger)

	users := rg.Group("/users")
	{

		users.GET("/search", deps.UserHandler.SearchUsers)
		users.GET("/:id", deps.UserHandler.GetUserByID)

		users.GET("/profile", authMiddleware, deps.UserHandler.GetProfile)
		users.PUT("/profile", authMiddleware, deps.UserHandler.UpdateProfile)

		users.POST("/profile/picture",
			authMiddleware,
			middleware.FileUploadMiddleware(5<<20, []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}),
			deps.UserHandler.UploadProfilePicture,
		)
	}
}
