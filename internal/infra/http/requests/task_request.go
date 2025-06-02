package requests

import (
	"time"

	"github.com/BohdanBoriak/boilerplate-go-back/internal/domain"
)

type TaskRequest struct {
	Title       string  `json:"title" validate:"required"`
	Description *string `json:"description"`
	Date        *int64  `json:"date"`
}

// Нова структура для запиту на оновлення статусу
type UpdateTaskStatusRequest struct {
	Status domain.TaskStatus `json:"status" validate:"required,oneof=NEW IN_PROGRESS COMPLETE"`
}

func (r TaskRequest) ToDomainModel() (interface{}, error) {
	var taskDate *time.Time
	if r.Date != nil {
		t := time.Unix(*r.Date, 0)
		taskDate = &t
	}

	return domain.Task{
		Title:       r.Title,
		Description: r.Description,
		Date:        taskDate,
	}, nil
}

// метод ToDomainModel для нової структури
func (r UpdateTaskStatusRequest) ToDomainModel() (interface{}, error) {
	return r.Status, nil
}
