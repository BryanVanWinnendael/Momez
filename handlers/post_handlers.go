package handlers

import (
	"fmt"
	"mime/multipart"
	"momez/dto"
	"momez/views/components"
	"momez/views/discover_views"
	"momez/views/favorite_views"
	"momez/views/post_views"
	"momez/views/tag_views"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type PostService interface {
	UploadPost(c echo.Context, fileHeader *multipart.FileHeader, caption string, username string, tag string) error
	GetPosts(c echo.Context, username string) ([]*dto.PostDto, error)
	GetUserPosts(c echo.Context, usernamePost string, tag string, username string) ([]*dto.PostDto, error)
	GetPost(c echo.Context, id string) (*dto.PostDto, error)
	EditPost(c echo.Context, id string, caption string, username string, selectedTag string) error
	DeletePost(c echo.Context, id string, username string) error
	GetTags(c echo.Context, username string) ([]string, error)
	AddTag(c echo.Context, username string, tag string) error
	GetsPostsPerTag(c echo.Context, tag string) (map[string]int, error)
	DeleteTag(c echo.Context, username string, tag string) error
	GetPostsWithTag(c echo.Context, tag string, username string) ([]*dto.PostDto, error)
	GetTop5RecentTagSearches(c echo.Context, username string) ([]string, error)
	FavoritePost(c echo.Context, id string, username string) error
	GetFavorites(c echo.Context, username string) ([]*dto.PostDto, error)
}

func NewPostHandler(ps PostService) *PostHandler {

	return &PostHandler{
		PostServices: ps,
	}
}

type PostHandler struct {
	PostServices PostService
}

func (ps *PostHandler) createPostHandler(c echo.Context) error {
	isError = false
	username := c.Get(username_key).(string)

	if c.Request().Method == "POST" {
		var selectedTag string

		tags := c.FormValue("tags")

		if tags == "add-new-tag" {
			selectedTag = strings.TrimSpace(c.FormValue("new-tag"))

			err := ps.PostServices.AddTag(c, username, selectedTag)
			if err != nil {
				setFlashmessages(c, "error", "Failed to add tag")
				return c.Redirect(http.StatusSeeOther, "/posts/upload")
			}
		} else if tags == "No tag selected" {
			selectedTag = ""
		} else {
			selectedTag = tags
		}

		caption := c.FormValue("caption")
		fileHeader, err := c.FormFile("file")
		if err != nil {
			setFlashmessages(c, "error", "Image is required")
			return c.Redirect(http.StatusSeeOther, "/posts/upload")
		}

		err = ps.PostServices.UploadPost(c, fileHeader, caption, username, selectedTag)
		if err != nil {
			setFlashmessages(c, "error", "Failed to create post")
			return c.Redirect(http.StatusSeeOther, "/posts/upload")
		}

		setFlashmessages(c, "success", "Post created successfully")
		return c.Redirect(http.StatusSeeOther, "/posts")
	}

	tags, err := ps.PostServices.GetTags(c, username)

	if err != nil {
		return err
	}

	return renderView(c, post_views.PostIndex(
		"| Create Post",
		c.Get(username_key).(string),
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		post_views.CreatePost(tags),
	))
}

func (ps *PostHandler) listPostHandler(c echo.Context) error {
	isError = false

	username := c.Get(username_key).(string)

	posts, err := ps.PostServices.GetPosts(c, username)
	if err != nil {
		return err
	}

	return renderView(c, post_views.PostIndex(
		"| Home",
		c.Get(username_key).(string),
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		post_views.PostList(posts, false),
	))
}

func (ps *PostHandler) getUserPostsHandler(c echo.Context) error {
	tag := c.QueryParam("tag")
	if tag != "" {
		return ps.GetUserPostsByTagHandler(c)
	}
	usernamePost := c.Param("username")
	activeUser := c.Get(username_key).(string)

	canEdit := usernamePost == activeUser

	isError = false

	posts, err := ps.PostServices.GetUserPosts(c, usernamePost, tag, activeUser)
	if err != nil {
		return err
	}

	tags, err := ps.PostServices.GetTags(c, usernamePost)
	if err != nil {
		return err
	}

	return renderView(c, post_views.PostIndex(
		"| User",
		c.Get(username_key).(string),
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		post_views.PostUser(posts, true, usernamePost, canEdit, tags),
	))
}

func (ps *PostHandler) GetUserPostsByTagHandler(c echo.Context) error {
	tag := c.QueryParam("tag")
	usernamePost := c.Param("username")
	activeUser := c.Get(username_key).(string)

	canEdit := usernamePost == activeUser

	isError = false

	posts, err := ps.PostServices.GetUserPosts(c, usernamePost, tag, activeUser)
	if err != nil {
		return err
	}

	return components.Posts(posts, canEdit).Render(c.Request().Context(), c.Response().Writer)
}

func (ps *PostHandler) editPostHandler(c echo.Context) error {
	isError = false

	if c.Request().Method == "POST" {
		username := c.Get(username_key).(string)
		id := c.FormValue("id")
		caption := c.FormValue("caption")

		var selectedTag string

		tags := c.FormValue("tags")

		if tags == "add-new-tag" {
			selectedTag = strings.TrimSpace(c.FormValue("new-tag"))

			err := ps.PostServices.AddTag(c, username, selectedTag)
			if err != nil {
				setFlashmessages(c, "error", "Failed to add tag")
				return c.Redirect(http.StatusSeeOther, "/posts/upload")
			}
		} else if tags == "No tag selected" {
			selectedTag = ""
		} else {
			selectedTag = tags
		}

		err := ps.PostServices.EditPost(c, id, caption, username, selectedTag)
		if err != nil {
			setFlashmessages(c, "error", "Failed to edit post")
			route := fmt.Sprintf("/posts/%s", id)
			return c.Redirect(http.StatusSeeOther, route)
		}

		setFlashmessages(c, "success", "Post edit successfully")
		route := fmt.Sprintf("/users/%s", username)
		return c.Redirect(http.StatusSeeOther, route)
	}

	postID := c.Param("id")

	post, err := ps.PostServices.GetPost(c, postID)
	if err != nil {
		setFlashmessages(c, "error", "Failed to get post")
		return c.Redirect(http.StatusSeeOther, "/posts")
	}

	tags, err := ps.PostServices.GetTags(c, c.Get(username_key).(string))
	if err != nil {
		return err
	}

	return renderView(c, post_views.PostIndex(
		"| Edit Post",
		c.Get(username_key).(string),
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		post_views.EditPost(post, tags),
	))
}

func (ps *PostHandler) deletePostHandler(c echo.Context) error {
	isError = false

	id := c.Param("id")
	username := c.Get(username_key).(string)

	err := ps.PostServices.DeletePost(c, id, username)
	if err != nil {
		setFlashmessages(c, "error", "Failed to delete post")
		route := fmt.Sprintf("/posts/%s", id)
		return c.Redirect(http.StatusSeeOther, route)
	}

	setFlashmessages(c, "success", "Post deleted successfully")
	route := fmt.Sprintf("/users/%s", username)
	return c.Redirect(http.StatusSeeOther, route)
}

func (ps *PostHandler) getUserEditTagsHandler(c echo.Context) error {
	isError = false
	username := c.Get(username_key).(string)

	tags, err := ps.PostServices.GetTags(c, username)
	if err != nil {
		return err
	}

	postsPerTag, err := ps.PostServices.GetsPostsPerTag(c, username)
	if err != nil {
		return err
	}

	return renderView(c, tag_views.TagIndex(
		"| Edit tags",
		c.Get(username_key).(string),
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		tag_views.TagList(tags, postsPerTag),
	))
}

func (ps *PostHandler) deleteTagHandler(c echo.Context) error {
	tag := c.QueryParam("tag")
	username := c.Get(username_key).(string)

	err := ps.PostServices.DeleteTag(c, username, tag)
	if err != nil {
		setFlashmessages(c, "error", "Failed to delete tag")
		return c.Redirect(http.StatusSeeOther, "/tags")
	}

	setFlashmessages(c, "success", "Tag deleted successfully")
	return c.Redirect(http.StatusSeeOther, "/tags/edit")
}

func (ps *PostHandler) addTagHandler(c echo.Context) error {
	tag := c.FormValue("tag")
	username := c.Get(username_key).(string)

	err := ps.PostServices.AddTag(c, username, tag)
	if err != nil {
		setFlashmessages(c, "error", "Failed to add tag")
		return c.Redirect(http.StatusSeeOther, "/tags/edit")
	}

	setFlashmessages(c, "success", "Tag added successfully")
	return c.Redirect(http.StatusSeeOther, "/tags/edit")
}

func (ps *PostHandler) getDiscoverHandler(c echo.Context) error {
	isError = false

	username := c.Get(username_key).(string)

	posts, err := ps.PostServices.GetPosts(c, username)
	if err != nil {
		return err
	}

	tags, err := ps.PostServices.GetTop5RecentTagSearches(c, username)
	if err != nil {
		return err
	}

	return renderView(c, discover_views.DiscoverIndex(
		"| Discover",
		c.Get(username_key).(string),
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		discover_views.DiscoverList(posts, tags),
	))
}

func (ps *PostHandler) getPostsWithTagHandler(c echo.Context) error {
	tag := c.FormValue("tag")
	username := c.Get(username_key).(string)

	posts, err := ps.PostServices.GetPostsWithTag(c, tag, username)
	if err != nil {
		return err
	}

	return components.Posts(posts, false).Render(c.Request().Context(), c.Response().Writer)
}

func (ps *PostHandler) favoritePostHandler(c echo.Context) error {
	username := c.Get(username_key).(string)
	id := c.Param("id")

	err := ps.PostServices.FavoritePost(c, id, username)
	if err != nil {
		return err
	}

	// get route from referer
	route := c.Request().Referer()
	usernamePost := ""

	if strings.Contains(route, "/users/") {
		usernamePost = strings.Split(route, "/")[4]
		posts, err := ps.PostServices.GetUserPosts(c, usernamePost, "", username)
		if err != nil {
			return err
		}
		canEdit := usernamePost == username
		return components.Posts(posts, canEdit).Render(c.Request().Context(), c.Response().Writer)
	} else if strings.Contains(route, "/favorites") {
		posts, err := ps.PostServices.GetFavorites(c, username)
		if err != nil {
			return err
		}
		canEdit := false // Assuming favorites cannot be edited
		return components.Posts(posts, canEdit).Render(c.Request().Context(), c.Response().Writer)
	}

	// If none of the conditions are met, return all posts
	posts, err := ps.PostServices.GetPosts(c, username)
	if err != nil {
		return err
	}
	return components.Posts(posts, false).Render(c.Request().Context(), c.Response().Writer)
}

func (ps *PostHandler) getFavoritesHandler(c echo.Context) error {
	isError = false

	username := c.Get(username_key).(string)

	posts, err := ps.PostServices.GetFavorites(c, username)
	if err != nil {
		return err
	}

	return renderView(c, favorite_views.FavoriteIndex(
		"| Favorites",
		c.Get(username_key).(string),
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		favorite_views.FavoriteList(posts, false),
	))
}
