package services

import (
	"fmt"
	"mime/multipart"
	"momez/db"
	"momez/dto"

	"github.com/google/uuid"

	"github.com/labstack/echo/v4"
)

func NewPostServices(database *db.Database) *PostServices {

	return &PostServices{
		db: database,
	}
}

type PostServices struct {
	db *db.Database
}

func (ps *PostServices) UploadPost(c echo.Context, fileHeader *multipart.FileHeader, caption string, username string, tag string) error {
	// Add an image to storage
	context := c.Request().Context()
	id := uuid.New().String()
	url, err := ps.db.UploadImageToFirebaseStorage(context, fileHeader, id)

	if err != nil {
		return err
	}

	err = ps.db.UploadPost(context, url, caption, username, id, tag)
	if err != nil {
		return err
	}

	return nil
}

func (ps *PostServices) GetPosts(c echo.Context, username string) ([]*dto.PostDto, error) {
	posts, err := ps.db.GetPosts(c.Request().Context(), username)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (ps *PostServices) GetUserPosts(c echo.Context, usernamePost string, tag string, username string) ([]*dto.PostDto, error) {
	posts, err := ps.db.GetUserPosts(c.Request().Context(), usernamePost, tag, username)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (ps *PostServices) GetPost(c echo.Context, id string) (*dto.PostDto, error) {
	post, err := ps.db.GetPost(c.Request().Context(), id)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (ps *PostServices) EditPost(c echo.Context, id string, caption string, username string, selectedTag string) error {
	context := c.Request().Context()

	hasPremission := ps.db.HasPostPremission(context, id, username)
	if !hasPremission {
		return fmt.Errorf("you don't have permission to delete this post")
	}

	err := ps.db.EditPost(context, id, caption, selectedTag)
	if err != nil {
		return err
	}
	return nil
}

func (ps *PostServices) DeletePost(c echo.Context, id string, username string) error {
	context := c.Request().Context()

	hasPremission := ps.db.HasPostPremission(context, id, username)
	if !hasPremission {
		return fmt.Errorf("you don't have permission to delete this post")
	}

	err := ps.db.DeletePost(context, id)
	if err != nil {
		return err
	}
	return nil
}

func (ps *PostServices) GetTags(c echo.Context, username string) ([]string, error) {
	tags, err := ps.db.GetUserTags(c.Request().Context(), username)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (ps *PostServices) AddTag(c echo.Context, username string, tag string) error {
	context := c.Request().Context()
	err := ps.db.AddTag(context, tag, username)
	if err != nil {
		return err
	}
	return nil
}

func (ps *PostServices) GetsPostsPerTag(c echo.Context, tag string) (map[string]int, error) {
	posts, err := ps.db.GetCountPostsPerTag(c.Request().Context(), tag)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (ps *PostServices) DeleteTag(c echo.Context, username string, tag string) error {
	context := c.Request().Context()
	err := ps.db.DeleteTag(context, tag, username)
	if err != nil {
		return err
	}
	return nil
}

func (ps *PostServices) GetPostsWithTag(c echo.Context, tag string, username string) ([]*dto.PostDto, error) {
	posts, err := ps.db.GetPostsWithTag(c.Request().Context(), tag, username)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (ps *PostServices) GetTop5RecentTagSearches(c echo.Context, username string) ([]string, error) {
	tags, err := ps.db.GetTop5RecentTagSearches(c.Request().Context(), username)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (ps *PostServices) FavoritePost(c echo.Context, id string, username string) error {
	context := c.Request().Context()
	err := ps.db.HandlePostFavorites(context, id, username)
	if err != nil {
		return err
	}
	return nil
}

func (ps *PostServices) GetFavorites(c echo.Context, username string) ([]*dto.PostDto, error) {
	posts, err := ps.db.GetFavoritePosts(c.Request().Context(), username)
	if err != nil {
		return nil, err
	}
	return posts, nil
}
