package dto

type PostDto struct {
	URL        string `json:"url"`
	Caption    string `json:"caption"`
	Username   string `json:"username"`
	DateString string `json:"date_string"`
	CreatedAt  string `json:"created_at"`
	ID         string `json:"id"`
	TAG        string `json:"tag"`
	FAVORITED  bool   `json:"favorited"`
}
