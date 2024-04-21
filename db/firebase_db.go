package db

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"momez/dto"
	"net/url"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	firebaseDb "firebase.google.com/go/v4/db"
)

func NewDB(storage *storage.Client, database *firebaseDb.Client) *Database {

	return &Database{
		storage:  storage,
		database: database,
	}
}

type Database struct {
	storage  *storage.Client
	database *firebaseDb.Client
	store    *firestore.Client
}

func getDateString() string {
	now := time.Now()
	months := [...]string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	day := now.Day()
	month := months[now.Month()-1] // Subtract 1 because month numbering starts from 1
	year := now.Year()

	formattedDate := fmt.Sprintf("%d %s", day, month)
	if year != 0 {
		formattedDate += fmt.Sprintf(" %d", year)
	}

	return formattedDate
}

func formatDate(dateString string) string {
	parts := strings.Split(dateString, " ")
	day, _ := strconv.Atoi(parts[0])
	monthStr := parts[1]
	yearStr := parts[2]

	// Get the current year
	currentYear := time.Now().Year()

	// Check if the year in the string is the current year
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return ""
	}

	// Remove year from the string if it's the current year
	if year == currentYear {
		dateString = fmt.Sprintf("%d %s", day, monthStr)
	}

	return dateString
}

func (db *Database) UploadPost(ctx context.Context, url string, caption string, username string, id string, tag string) error {
	createdAt := time.Now().Format(time.RFC3339)
	dateString := getDateString()

	post := dto.PostDto{URL: url, Caption: caption, Username: username, DateString: dateString, CreatedAt: createdAt, ID: id, TAG: tag}

	ref := db.database.NewRef("posts")

	err := ref.Child(id).Set(ctx, post)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) GetPosts(ctx context.Context, username string) ([]*dto.PostDto, error) {
	ref := db.database.NewRef("posts")

	var posts map[string]dto.PostDto
	if err := ref.Get(ctx, &posts); err != nil {
		return nil, err
	}

	var postDtos []*dto.PostDto
	for _, post := range posts {
		post.DateString = formatDate(post.DateString)
		postDtos = append(postDtos, &post)
	}

	favorites, err := db.GetFavoritePosts(ctx, username)
	if err != nil {
		return nil, err
	}

	for _, post := range postDtos {
		if db.IsPostFavovited(ctx, post.ID, username, favorites) {
			post.FAVORITED = true
		}
	}

	sort.Slice(postDtos, func(i, j int) bool {
		time1, _ := time.Parse(time.RFC3339, postDtos[i].CreatedAt)
		time2, _ := time.Parse(time.RFC3339, postDtos[j].CreatedAt)
		return time2.Before(time1)
	})

	return postDtos, nil
}

func (db *Database) UploadImageToFirebaseStorage(ctx context.Context, fileHeader *multipart.FileHeader, id string) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	filename := "images/" + id

	// Create a storage bucket object
	bucketName := os.Getenv("BUCKET_NAME")
	bucket := db.storage.Bucket(bucketName)

	// Create an object handle
	obj := bucket.Object(filename)

	// Create a writer for the object
	writer := obj.NewWriter(ctx)
	writer.ObjectAttrs.Metadata = map[string]string{
		"firebaseStorageDownloadTokens": "public",
	}

	// Copy the file data to the object's writer
	_, err = io.Copy(writer, file)
	if err != nil {
		return "", err
	}
	defer writer.Close()

	encodedFilename := url.PathEscape(filename)
	// Get the public URL of the uploaded file
	url := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media&token=public", os.Getenv("BUCKET_NAME"), encodedFilename)
	return url, nil
}

func (db *Database) GetUserPosts(ctx context.Context, usernamePost string, tag string, username string) ([]*dto.PostDto, error) {
	ref := db.database.NewRef("posts")

	var posts map[string]dto.PostDto
	if err := ref.OrderByChild("username").EqualTo(usernamePost).Get(ctx, &posts); err != nil {
		return nil, err
	}

	if tag != "" && tag != "all" {
		filteredPosts := make(map[string]dto.PostDto)
		for _, post := range posts {
			if post.TAG == tag {
				filteredPosts[post.ID] = post
			}
		}
		posts = filteredPosts
	}

	favorites, err := db.GetFavoritePosts(ctx, username)
	if err != nil {
		return nil, err
	}

	var postDtos []*dto.PostDto
	for _, post := range posts {
		post.DateString = formatDate(post.DateString)
		if db.IsPostFavovited(ctx, post.ID, username, favorites) {
			post.FAVORITED = true
		}
		postDtos = append(postDtos, &post)
	}

	sort.Slice(postDtos, func(i, j int) bool {
		time1, _ := time.Parse(time.RFC3339, postDtos[i].CreatedAt)
		time2, _ := time.Parse(time.RFC3339, postDtos[j].CreatedAt)
		return time2.Before(time1)
	})

	return postDtos, nil
}

func (db *Database) GetPost(ctx context.Context, id string) (*dto.PostDto, error) {
	ref := db.database.NewRef("posts")

	var posts map[string]dto.PostDto
	if err := ref.OrderByChild("id").EqualTo(id).Get(ctx, &posts); err != nil {
		return nil, err
	}

	var post *dto.PostDto
	for _, p := range posts {
		post = &p
	}

	return post, nil
}

func (db *Database) EditPost(ctx context.Context, id string, caption string, selectedTag string) error {
	post, err := db.GetPost(ctx, id)
	if err != nil {
		return err
	}

	post.Caption = caption
	post.TAG = selectedTag

	ref := db.database.NewRef("posts/" + id)
	err = ref.Set(ctx, post)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) deleteImageFromFirebaseStorage(ctx context.Context, filename string) error {
	bucketName := os.Getenv("BUCKET_NAME")
	bucket := db.storage.Bucket(bucketName)

	obj := bucket.Object(filename)
	err := obj.Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) DeletePost(ctx context.Context, id string) error {
	ref := db.database.NewRef("posts/" + id)
	err := ref.Delete(ctx)
	if err != nil {
		return err
	}

	filename := "images/" + id
	err = db.deleteImageFromFirebaseStorage(ctx, filename)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) SetProfileImage(ctx context.Context, fileHeader *multipart.FileHeader, username string) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	filename := "profile_images/" + username

	bucketName := os.Getenv("BUCKET_NAME")
	bucket := db.storage.Bucket(bucketName)

	obj := bucket.Object(filename)

	writer := obj.NewWriter(ctx)
	writer.ObjectAttrs.Metadata = map[string]string{
		"firebaseStorageDownloadTokens": "public",
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		return "", err
	}
	defer writer.Close()

	encodedFilename := url.PathEscape(filename)
	url := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media&token=public", os.Getenv("BUCKET_NAME"), encodedFilename)
	return url, nil
}

func (db *Database) UploadDefaultProfileImage(ctx context.Context, username string) (string, error) {
	file, err := os.Open("static/img/profile.svg")
	if err != nil {
		return "", err
	}
	defer file.Close()

	filename := "profile_images/" + username

	bucketName := os.Getenv("BUCKET_NAME")
	bucket := db.storage.Bucket(bucketName)

	obj := bucket.Object(filename)

	writer := obj.NewWriter(ctx)
	writer.ObjectAttrs.ContentType = "image/svg+xml"
	writer.ObjectAttrs.Metadata = map[string]string{
		"firebaseStorageDownloadTokens": "public",
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		writer.Close()
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	encodedFilename := url.PathEscape(filename)

	url := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media&token=public", os.Getenv("BUCKET_NAME"), encodedFilename)
	return url, nil
}

func (db *Database) SetBannerImage(ctx context.Context, fileHeader *multipart.FileHeader, username string) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	filename := "banner_images/" + username

	bucketName := os.Getenv("BUCKET_NAME")
	bucket := db.storage.Bucket(bucketName)

	obj := bucket.Object(filename)

	writer := obj.NewWriter(ctx)
	writer.ObjectAttrs.Metadata = map[string]string{
		"firebaseStorageDownloadTokens": "public",
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		return "", err
	}
	defer writer.Close()

	encodedFilename := url.PathEscape(filename)
	url := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media&token=public", os.Getenv("BUCKET_NAME"), encodedFilename)
	return url, nil
}

func (db *Database) UploadDefaultBannerImage(ctx context.Context, username string) (string, error) {
	file, err := os.Open("static/img/banner.svg")
	if err != nil {
		return "", err
	}
	defer file.Close()

	filename := "banner_images/" + username

	bucketName := os.Getenv("BUCKET_NAME")
	bucket := db.storage.Bucket(bucketName)

	obj := bucket.Object(filename)

	writer := obj.NewWriter(ctx)
	writer.ObjectAttrs.ContentType = "image/svg+xml"
	writer.ObjectAttrs.Metadata = map[string]string{
		"firebaseStorageDownloadTokens": "public",
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		writer.Close()
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	encodedFilename := url.PathEscape(filename)

	url := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media&token=public", os.Getenv("BUCKET_NAME"), encodedFilename)
	return url, nil
}

func (db *Database) HasPostPremission(ctx context.Context, id string, username string) bool {
	post, err := db.GetPost(ctx, id)
	if err != nil {
		return false
	}

	if post.Username != username {
		return false
	}

	return true
}

func (db *Database) GetUserTags(ctx context.Context, username string) ([]string, error) {
	ref := db.database.NewRef("tags/" + username)

	var tags []string
	if err := ref.Get(ctx, &tags); err != nil {
		return nil, err
	}

	return tags, nil
}

func (db *Database) AddTag(ctx context.Context, tag string, username string) error {
	tags, err := db.GetUserTags(ctx, username)
	if err != nil {
		return err
	}

	if strings.TrimSpace(tag) == "" {
		return nil
	}

	// Check if the tag already exists
	for _, t := range tags {
		if t == tag {
			return nil
		}
	}

	tags = append(tags, tag)

	ref := db.database.NewRef("tags/" + username)
	err = ref.Set(ctx, tags)
	if err != nil {
		return err
	}

	refAll := db.database.NewRef("all_tags")
	var allTags []string
	if err := refAll.Get(ctx, &allTags); err != nil {
		return err
	}

	// Check if the tag already exists in the all_tags list
	found := false
	for _, t := range allTags {
		if t == tag {
			found = true
			break
		}
	}

	if !found {
		allTags = append(allTags, tag)
		ref := db.database.NewRef("all_tags")
		err = ref.Set(ctx, allTags)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) GetCountPostsPerTag(ctx context.Context, username string) (map[string]int, error) {
	ref := db.database.NewRef("posts")

	var posts map[string]dto.PostDto
	if err := ref.OrderByChild("username").EqualTo(username).Get(ctx, &posts); err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, post := range posts {
		counts[post.TAG]++
	}

	return counts, nil
}

func (db *Database) DeleteTag(ctx context.Context, tag string, username string) error {
	tags, err := db.GetUserTags(ctx, username)
	if err != nil {
		return err
	}

	newTags := make([]string, 0)
	for _, t := range tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}

	ref := db.database.NewRef("tags/" + username)
	err = ref.Set(ctx, newTags)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) AddFoundTagToRecentCounts(ctx context.Context, tag string, username string) {
	ref := db.database.NewRef("recent_searched_tags/" + username)

	// if the tag already exists, increment the count
	var count int
	if err := ref.Child(tag).Get(ctx, &count); err != nil {
		count = 0
	}

	count++
	ref.Child(tag).Set(ctx, count)
}

func (db *Database) GetPostsWithTag(ctx context.Context, tag string, username string) ([]*dto.PostDto, error) {
	ref := db.database.NewRef("posts")

	var posts map[string]dto.PostDto
	if err := ref.Get(ctx, &posts); err != nil {
		return nil, err
	}

	var postDtos []*dto.PostDto
	for _, post := range posts {
		if strings.Contains(strings.ToLower(post.TAG), strings.ToLower(tag)) {
			if tag != "" {
				db.AddFoundTagToRecentCounts(ctx, post.TAG, username)
			}

			post.DateString = formatDate(post.DateString)
			postDtos = append(postDtos, &post)
		}
	}

	sort.Slice(postDtos, func(i, j int) bool {
		time1, _ := time.Parse(time.RFC3339, postDtos[i].CreatedAt)
		time2, _ := time.Parse(time.RFC3339, postDtos[j].CreatedAt)
		return time2.Before(time1)
	})

	return postDtos, nil
}

func (db *Database) GetTop5RecentTagSearches(ctx context.Context, username string) ([]string, error) {
	ref := db.database.NewRef("recent_searched_tags/" + username)

	var tags map[string]int
	if err := ref.Get(ctx, &tags); err != nil {
		return nil, err
	}

	type tagCount struct {
		Tag   string
		Count int
	}

	var tagCounts []tagCount
	for tag, count := range tags {
		tagCounts = append(tagCounts, tagCount{Tag: tag, Count: count})
	}

	sort.Slice(tagCounts, func(i, j int) bool {
		return tagCounts[i].Count > tagCounts[j].Count
	})

	var top5Tags []string
	for i, tagCount := range tagCounts {
		if i == 5 {
			break
		}
		top5Tags = append(top5Tags, tagCount.Tag)
	}

	return top5Tags, nil
}

func (db *Database) HandlePostFavorites(ctx context.Context, postID string, username string) error {
	ref := db.database.NewRef("favorites/" + username)

	var favorites []string
	if err := ref.Get(ctx, &favorites); err != nil {
		return err
	}

	// Check if the post is already in the favorites list
	found := false
	for _, id := range favorites {
		if id == postID {
			found = true
			break
		}
	}

	if found {
		// Remove the post from the favorites list
		newFavorites := make([]string, 0)
		for _, id := range favorites {
			if id != postID {
				newFavorites = append(newFavorites, id)
			}
		}

		favorites = newFavorites
	} else {
		// Add the post to the favorites list
		favorites = append(favorites, postID)
	}

	err := ref.Set(ctx, favorites)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) GetFavoritePosts(ctx context.Context, username string) ([]*dto.PostDto, error) {
	ref := db.database.NewRef("posts")

	var posts map[string]dto.PostDto
	if err := ref.Get(ctx, &posts); err != nil {
		return nil, err
	}

	refFavorites := db.database.NewRef("favorites/" + username)

	var favorites []string
	if err := refFavorites.Get(ctx, &favorites); err != nil {
		return nil, err
	}

	var favoritePosts []*dto.PostDto
	for _, post := range posts {
		for _, id := range favorites {
			if post.ID == id {
				post.DateString = formatDate(post.DateString)
				post.FAVORITED = true
				favoritePosts = append(favoritePosts, &post)
			}
		}
	}

	var tags []string
	for _, post := range favoritePosts {
		if !slices.Contains(tags, post.TAG) {
			tags = append(tags, post.TAG)
		}
	}

	sort.Slice(favoritePosts, func(i, j int) bool {
		time1, _ := time.Parse(time.RFC3339, favoritePosts[i].CreatedAt)
		time2, _ := time.Parse(time.RFC3339, favoritePosts[j].CreatedAt)
		return time2.Before(time1)
	})

	return favoritePosts, nil
}

func (db *Database) IsPostFavovited(ctx context.Context, postID string, username string, favorites []*dto.PostDto) bool {
	for _, post := range favorites {
		if post.ID == postID {
			return true
		}
	}

	return false
}
