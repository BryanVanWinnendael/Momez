package services

import (
	"mime/multipart"
	"momez/db"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func NewUserServices(u User, uStore db.Store, database *db.Database) *UserServices {

	return &UserServices{
		User:      u,
		UserStore: uStore,
		db:        database,
	}
}

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type UserServices struct {
	User      User
	UserStore db.Store
	db        *db.Database
}

func (us *UserServices) CreateUser(u User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 8)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users(email, password, username) VALUES($1, $2, $3)`

	_, err = us.UserStore.Db.Exec(
		stmt,
		u.Email,
		string(hashedPassword),
		u.Username,
	)

	return err
}

func (us *UserServices) CheckEmail(email string) (User, error) {

	query := `SELECT id, email, password, username FROM users
		WHERE email = ?`

	stmt, err := us.UserStore.Db.Prepare(query)
	if err != nil {
		return User{}, err
	}

	defer stmt.Close()

	us.User.Email = email
	err = stmt.QueryRow(
		us.User.Email,
	).Scan(
		&us.User.ID,
		&us.User.Email,
		&us.User.Password,
		&us.User.Username,
	)
	if err != nil {
		return User{}, err
	}

	return us.User, nil
}

func (us *UserServices) SetDefaultProfileImage(c echo.Context, username string) (string, error) {
	context := c.Request().Context()

	url, err := us.db.UploadDefaultProfileImage(context, username)
	if err != nil {
		return "", err
	}

	return url, nil
}

func (us *UserServices) SetDefaultBannerImage(c echo.Context, username string) (string, error) {
	context := c.Request().Context()

	url, err := us.db.UploadDefaultBannerImage(context, username)
	if err != nil {
		return "", err
	}

	return url, nil
}

func (us *UserServices) SetProfileImage(c echo.Context, username string, fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	context := c.Request().Context()

	url, err := us.db.SetProfileImage(context, fileHeader, username)
	if err != nil {
		return "", err
	}

	return url, nil
}

func (us *UserServices) SetBannerImage(c echo.Context, username string, fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	context := c.Request().Context()

	url, err := us.db.SetBannerImage(context, fileHeader, username)
	if err != nil {
		return "", err
	}

	return url, nil
}
