package database

import (
	"time"

	"github.com/BohdanBoriak/boilerplate-go-back/internal/domain"
	"github.com/upper/db/v4"
)

const TasksTableName = "tasks"

type task struct {
	Id          uint64            `db:"id,omitempty"`
	UserId      uint64            `db:"user_id"`
	Title       string            `db:"title"`
	Description *string           `db:"description"`
	Date        *time.Time        `db:"date"`
	Status      domain.TaskStatus `db:"status"`
	CreatedDate time.Time         `db:"created_date"`
	UpdatedDate time.Time         `db:"updated_date"`
	DeletedDate *time.Time        `db:"deleted_date"`
}

type TaskRepository interface {
	Save(t domain.Task) (domain.Task, error)
	Find(id uint64) (domain.Task, error)
	// Змінений FindAllTasks
	FindAllTasks(uId uint64, status *domain.TaskStatus, date *time.Time) ([]domain.Task, error)
	Update(t domain.Task) (domain.Task, error)
	Delete(id uint64) error
	// UpdateStatus
	UpdateStatus(taskId uint64, status domain.TaskStatus) error
}

type taskRepository struct {
	coll db.Collection
	sess db.Session
}

func NewTaskRepository(sess db.Session) TaskRepository {
	return taskRepository{
		coll: sess.Collection(TasksTableName),
		sess: sess,
	}
}

func (r taskRepository) Save(t domain.Task) (domain.Task, error) {
	tsk := r.mapDomainToModel(t)
	// Встановлення CreatedDate та UpdatedDate при створенні
	if tsk.CreatedDate.IsZero() {
		tsk.CreatedDate = time.Now()
	}
	tsk.UpdatedDate = time.Now()

	err := r.coll.InsertReturning(&tsk)
	if err != nil {
		return domain.Task{}, err
	}

	t = r.mapModelToDomain(tsk)
	return t, nil
}

func (r taskRepository) Find(id uint64) (domain.Task, error) {
	var t task

	err := r.coll.Find(db.Cond{
		"id":           id,
		"deleted_date": nil,
	}).One(&t)
	if err != nil {
		return domain.Task{}, err
	}

	return r.mapModelToDomain(t), nil
}

// відредагован FindAllTasks
func (r taskRepository) FindAllTasks(uId uint64, status *domain.TaskStatus, date *time.Time) ([]domain.Task, error) {
	var ts []task

	conditions := db.Cond{
		"user_id":      uId,
		"deleted_date": nil,
	}

	// умови для фільтрації по статусу
	if status != nil && *status != "" {
		conditions["status"] = *status
	}

	// умови для фільтрації по даті
	if date != nil {
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond) // Кінець дня (23:59:59.999...)
		conditions["date >="] = startOfDay
		conditions["date <="] = endOfDay
	}

	// сортування за датою
	err := r.coll.Find(conditions).OrderBy("-created_date").All(&ts)
	if err != nil {
		return nil, err
	}
	return r.mapModelToDomainCollection(ts), nil
}

func (r taskRepository) Update(t domain.Task) (domain.Task, error) {
	tsk := r.mapDomainToModel(t)
	tsk.UpdatedDate = time.Now()
	err := r.coll.Find(db.Cond{"id": tsk.Id, "deleted_date": nil}).Update(&tsk)
	if err != nil {
		return domain.Task{}, err
	}
	return r.mapModelToDomain(tsk), nil
}

func (r taskRepository) Delete(id uint64) error {
	return r.coll.Find(db.Cond{"id": id, "deleted_date": nil}).Update(map[string]interface{}{"deleted_date": time.Now()})
}

// UpdateStatus
func (r taskRepository) UpdateStatus(taskId uint64, status domain.TaskStatus) error {
	return r.coll.Find(db.Cond{"id": taskId, "deleted_date": nil}).Update(map[string]interface{}{
		"status":       status,
		"updated_date": time.Now(),
	})
}

func (r taskRepository) mapDomainToModel(t domain.Task) task {
	return task{
		Id:          t.Id,
		UserId:      t.UserId,
		Title:       t.Title,
		Description: t.Description,
		Date:        t.Date,
		Status:      t.Status,
		CreatedDate: t.CreatedDate,
		UpdatedDate: t.UpdatedDate,
		DeletedDate: t.DeletedDate,
	}
}

func (r taskRepository) mapModelToDomain(t task) domain.Task {
	return domain.Task{
		Id:          t.Id,
		UserId:      t.UserId,
		Title:       t.Title,
		Description: t.Description,
		Date:        t.Date,
		Status:      t.Status,
		CreatedDate: t.CreatedDate,
		UpdatedDate: t.UpdatedDate,
		DeletedDate: t.DeletedDate,
	}
}

func (r taskRepository) mapModelToDomainCollection(ts []task) []domain.Task {
	tasks := make([]domain.Task, len(ts))
	for i, t := range ts {
		tasks[i] = r.mapModelToDomain(t)
	}
	return tasks
}
