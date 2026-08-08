package response

import "github.com/Rivaldz/my-template-go/internal/entity"

// CommentList -.
type CommentList struct {
	Comments []entity.Comment `json:"comments"`
	Total    int              `json:"total" example:"42"`
} // @name v1.CommentList
