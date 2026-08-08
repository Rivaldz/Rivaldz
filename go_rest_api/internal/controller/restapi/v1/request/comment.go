package request

// CreateComment -.
type CreateComment struct {
	Body string `json:"body" validate:"required,max=2000" example:"Nice post!"`
} // @name v1.CreateComment

// UpdateComment -.
type UpdateComment struct {
	Body string `json:"body" validate:"required,max=2000" example:"Updated comment"`
} // @name v1.UpdateComment
