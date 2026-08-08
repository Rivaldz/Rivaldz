package response

import "github.com/Rivaldz/my-template-go/internal/entity"

// TaskList -.
type TaskList struct {
	Tasks []entity.Task `json:"tasks"`
	Total int           `json:"total"`
}

// DeleteStatus -.
type DeleteStatus struct {
	Status string `json:"status"`
}
