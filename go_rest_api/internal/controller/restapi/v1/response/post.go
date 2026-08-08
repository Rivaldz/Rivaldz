package response

import "github.com/Rivaldz/my-template-go/internal/entity"

// PostList -.
type PostList struct {
	Posts []entity.Post `json:"posts"`
	Total int           `json:"total" example:"42"`
} // @name v1.PostList
