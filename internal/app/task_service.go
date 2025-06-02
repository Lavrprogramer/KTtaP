package app

import (
	"log"
	"time"

	"github.com/BohdanBoriak/boilerplate-go-back/internal/domain"
	"github.com/BohdanBoriak/boilerplate-go-back/internal/infra/database"
)

type TaskService interface {
	Save(t domain.Task) (domain.Task, error)
	Find(id uint64) (interface{}, error)
	// Змінено сигнатура методу FindAll
	FindAll(uId uint64, status *domain.TaskStatus, date *time.Time) ([]domain.Task, error)
	Update(t domain.Task) (domain.Task, error)
	Delete(id uint64) error
	//UpdateStatus
	UpdateStatus(taskId uint64, status domain.TaskStatus) (domain.Task, error)
}

type taskService struct {
	taskRepo database.TaskRepository
}

func NewTaskService(tr database.TaskRepository) TaskService {
	return taskService{
		taskRepo: tr,
	}
}

func (s taskService) Save(t domain.Task) (domain.Task, error) {
	//встановлення дати
	if t.CreatedDate.IsZero() {
		t.CreatedDate = time.Now()
	}
	t.UpdatedDate = time.Now()
	task, err := s.taskRepo.Save(t)
	if err != nil {
		log.Printf("taskService.Save(s.taskRepo.Save): %s", err)
		return domain.Task{}, err
	}
	return task, nil
}

func (s taskService) Find(id uint64) (interface{}, error) {
	task, err := s.taskRepo.Find(id)
	if err != nil {
		log.Printf("taskService.Find(s.taskRepo.Find): %s", err)
		return domain.Task{}, err
	}

	return task, nil
}

// оновлено FindAll
func (s taskService) FindAll(uId uint64, status *domain.TaskStatus, date *time.Time) ([]domain.Task, error) {
	tasks, err := s.taskRepo.FindAllTasks(uId, status, date)
	if err != nil {
		log.Printf("taskService.FindAll(s.taskRepo.FindAllTasks): %s", err)
		return nil, err
	}

	return tasks, nil
}

func (s taskService) Update(t domain.Task) (domain.Task, error) {
	// Встановлення часу UpdatedDate перед оновленням
	t.UpdatedDate = time.Now()
	task, err := s.taskRepo.Update(t)
	if err != nil {
		log.Printf("taskService.Update(s.taskRepo.Update): %s", err)
		return domain.Task{}, err
	}

	return task, nil
}

func (s taskService) Delete(id uint64) error {
	err := s.taskRepo.Delete(id)
	if err != nil {
		log.Printf("taskService.Delete(s.taskRepo.Delete): %s", err)
		return err
	}

	return nil
}

// UpdateStatus
func (s taskService) UpdateStatus(taskId uint64, status domain.TaskStatus) (domain.Task, error) {
	err := s.taskRepo.UpdateStatus(taskId, status)
	if err != nil {
		log.Printf("taskService.UpdateStatus(UpdateStatus): %s", err)
		return domain.Task{}, err
	}
	updatedTask, findErr := s.taskRepo.Find(taskId)
	if findErr != nil {
		log.Printf("taskService.UpdateStatus(Find after update): %s", findErr)
		return domain.Task{}, findErr
	}
	return updatedTask, nil
}
