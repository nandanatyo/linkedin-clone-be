package routes

import (
	"github.com/gin-gonic/gin"
	"linked-clone/internal/middleware"
)

func PostRoutes(rg *gin.RouterGroup, deps *Dependencies) {
	authMiddleware := middleware.AuthMiddleware(deps.JWTService, deps.Logger)

	posts := rg.Group("/posts")
	{

		posts.GET("/:id", deps.PostHandler.GetPost)
		posts.GET("/user/:user_id", deps.PostHandler.GetUserPosts)
		posts.GET("/:id/comments", deps.PostHandler.GetComments)

		posts.GET("", authMiddleware, deps.PostHandler.GetFeed)
		posts.PUT("/:id", authMiddleware, deps.PostHandler.UpdatePost)
		posts.DELETE("/:id", authMiddleware, deps.PostHandler.DeletePost)
		posts.POST("/:id/like", authMiddleware, deps.PostHandler.LikePost)
		posts.DELETE("/:id/like", authMiddleware, deps.PostHandler.UnlikePost)
		posts.POST("/:id/comments", authMiddleware, deps.PostHandler.AddComment)

		posts.POST("",
			authMiddleware,
			middleware.FileUploadMiddleware(10<<20, []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}),
			deps.PostHandler.CreatePost,
		)
	}
}
