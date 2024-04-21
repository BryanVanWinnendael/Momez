package handlers

import "github.com/labstack/echo/v4"

var (
	fromProtected bool = false
	isError       bool = false
)

func SetupRoutes(e *echo.Echo, ah *AuthHandler, ph *PostHandler) {
	e.GET("/", ah.loginHandler)
	e.POST("/", ah.loginHandler)
	e.GET("/login", ah.loginHandler)
	e.POST("/login", ah.loginHandler)
	e.GET("/register", ah.registerHandler)
	e.POST("/register", ah.registerHandler)

	protectedGroup := e.Group("/", ah.authMiddleware)

	protectedGroup.GET("posts", ph.listPostHandler)
	protectedGroup.GET("posts/upload", ph.createPostHandler)
	protectedGroup.POST("posts/upload", ph.createPostHandler)
	protectedGroup.GET("posts/:id", ph.editPostHandler)
	protectedGroup.POST("posts/:id", ph.editPostHandler)
	protectedGroup.DELETE("posts/:id", ph.deletePostHandler)
	protectedGroup.POST("posts/:id/favorite", ph.favoritePostHandler) // favorite from home page

	protectedGroup.GET("users/:username", ph.getUserPostsHandler)
	protectedGroup.POST("users/:username/profile-image", ah.setProfileImageHandler)
	protectedGroup.GET("users/:username/profile-image", ph.getUserPostsHandler)
	protectedGroup.POST("users/:username/banner-image", ah.setBannerImageHandler)
	protectedGroup.GET("users/:username/banner-image", ph.getUserPostsHandler)
	protectedGroup.GET("users/:username/tags", ph.GetUserPostsByTagHandler)

	protectedGroup.GET("tags/edit", ph.getUserEditTagsHandler)
	protectedGroup.DELETE("tags/edit", ph.deleteTagHandler)
	protectedGroup.POST("tags/edit", ph.addTagHandler)

	protectedGroup.GET("discover", ph.getDiscoverHandler)
	protectedGroup.POST("discover", ph.getPostsWithTagHandler)

	protectedGroup.GET("favorites", ph.getFavoritesHandler)

	protectedGroup.POST("logout", ah.logoutHandler)
}
