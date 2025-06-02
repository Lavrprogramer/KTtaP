package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/BohdanBoriak/boilerplate-go-back/internal/app"
	"github.com/BohdanBoriak/boilerplate-go-back/internal/domain"
	"github.com/BohdanBoriak/boilerplate-go-back/internal/infra/http/requests"
	"github.com/BohdanBoriak/boilerplate-go-back/internal/infra/http/resources"
)

type TaskController struct {
	taskService app.TaskService
}

func NewTaskController(ts app.TaskService) TaskController {
	return TaskController{
		taskService: ts,
	}
}

func (c TaskController) Save() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		task, err := requests.Bind(r, requests.TaskRequest{}, domain.Task{})
		if err != nil {
			log.Printf("TaskController: %s", err)
			BadRequest(w, err)
			return
		}

		task.Status = domain.TaskNew
		user := r.Context().Value(UserKey).(domain.User)
		task.UserId = user.Id

		savedTask, err := c.taskService.Save(task)
		if err != nil {
			log.Printf("TaskController: %s", err)
			InternalServerError(w, err)
			return
		}

		var taskDto resources.TaskDto
		Created(w, taskDto.DomainToDto(savedTask))
	}
}

func (c TaskController) Find() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		task := r.Context().Value(TaskKey).(domain.Task)
		// перенесено в PathObjectMiddleware
		// user := r.Context().Value(UserKey).(domain.User)
		// if task.UserId != user.Id {
		// 	err := errors.New("access denied")
		// 	Forbidden(w, err)
		// 	return
		// }

		var taskDto resources.TaskDto
		Success(w, taskDto.DomainToDto(task))
	}
}

func (c TaskController) FindAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserKey).(domain.User)
		//
		query := r.URL.Query()
		var statusFilter *domain.TaskStatus
		statusStr := query.Get("status")
		if statusStr != "" {
			s := domain.TaskStatus(statusStr)
			// перевірка значення статусу
			if s == domain.TaskNew || s == domain.TaskInProgress || s == domain.TaskComplete {
				statusFilter = &s
			} else {
				BadRequest(w, errors.New("invalid status filter value"))
				return
			}
		}
		var dateFilter *time.Time
		dateStr := query.Get("date")
		if dateStr != "" {
			parsedDate, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				log.Printf("TaskController %s", err)
				BadRequest(w, errors.New("invalid date format"))
				return
			}
			dateFilter = &parsedDate
		}

		tasks, err := c.taskService.FindAll(user.Id, statusFilter, dateFilter)
		if err != nil {
			log.Printf("TaskController.FindAll(c.taskService.FindAll): %s", err)
			InternalServerError(w, err)
			return
		}

		var taskDto resources.TaskDto
		tasksDto := taskDto.DomainToDtoCollection(tasks)
		Success(w, tasksDto)
	}
}
func (c TaskController) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskUpdates, err := requests.Bind(r, requests.TaskRequest{}, domain.Task{})
		if err != nil {
			log.Printf("TaskController: %s", err)
			BadRequest(w, err)
			return
		}

		existingTask := r.Context().Value(TaskKey).(domain.Task)

		// перенесено в PathObjectMiddleware
		/* 	taskExists := r.Context().Value(TaskKey).(domain.Task)
		if taskExists.UserId != user.Id {
			err = errors.New("Acces denied")
			Forbidden(w, err)
			return
		} */

		existingTask.Title = taskUpdates.Title
		existingTask.Description = taskUpdates.Description
		existingTask.Date = taskUpdates.Date

		updatedTask, err := c.taskService.Update(existingTask)
		if err != nil {
			log.Printf("TaskController: %s", err)
			InternalServerError(w, err)
			return
		}

		var taskDto resources.TaskDto
		Success(w, taskDto.DomainToDto(updatedTask))
	}
}

func (c TaskController) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		task := r.Context().Value(TaskKey).(domain.Task)
		//  PathObjectMiddleware
		// user := r.Context().Value(UserKey).(domain.User)
		// if task.UserId != user.Id {
		// 	err := errors.New("access denied")
		// 	Forbidden(w, err)
		// 	return
		// }
		err := c.taskService.Delete(task.Id)
		if err != nil {
			log.Printf("TaskController.Delete: c.taskService.Delete: %s", err)
			InternalServerError(w, err)
			return
		}
		noContent(w)
	}
}

// новий метод для оновлення статусу
func (c TaskController) UpdateStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskFromCtx := r.Context().Value(TaskKey).(domain.Task)
		var req requests.UpdateTaskStatusRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("TaskController.UpdateStatus: json.Decode: %s", err)
			BadRequest(w, errors.New("invalid request body for status update"))
			return
		}

		if req.Status != domain.TaskNew && req.Status != domain.TaskInProgress && req.Status != domain.TaskComplete {
			BadRequest(w, errors.New("invalid task status value"))
			return
		}

		updatedTask, err := c.taskService.UpdateStatus(taskFromCtx.Id, req.Status)
		if err != nil {
			log.Printf("TaskController.UpdateStatus: c.taskService.UpdateStatus: %s", err)
			InternalServerError(w, err)
			return
		}

		var taskDto resources.TaskDto
		Success(w, taskDto.DomainToDto(updatedTask))
	}
}
