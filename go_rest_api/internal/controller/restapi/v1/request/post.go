package request

// CreatePost -.
type CreatePost struct {
	Title string `json:"title" validate:"required,max=255" example:"My first post"`
	Body  string `json:"body"  validate:"required,max=10000" example:"Post content"`
} // @name v1.CreatePost

// UpdatePost -.
type UpdatePost struct {
	Title string `json:"title" validate:"required,max=255" example:"Updated post title"`
	Body  string `json:"body"  validate:"required,max=10000" example:"Updated post content"`
} // @name v1.UpdatePost
