package task

import (
	"errors"
	"sync"
	"TaskScheduler/internal/entities"
)

var ErrTaskNotFound = errors.New("task not found")

type Repository interface {
	Create(task *entities.Task) error
	Get(id string) (*entities.Task, error)
	Update(id string, updateFn func(t *entities.Task)) (*entities.Task, error)
}

type inMemoryRepository struct {
	mu    sync.RWMutex
	tasks map[string]*entities.Task
}

func NewInMemoryRepository() Repository {
	return &inMemoryRepository{
		tasks: make(map[string]*entities.Task),
	}
}

func (r *inMemoryRepository) Create(task *entities.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *inMemoryRepository) Get(id string) (*entities.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, exists := r.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}
	// return a copy to prevent external modification
	taskCopy := *task
	return &taskCopy, nil
}

func (r *inMemoryRepository) Update(id string, updateFn func(t *entities.Task)) (*entities.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	task, exists := r.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}
	updateFn(task)
	// return a copy to prevent external modification of repo state
	taskCopy := *task
	return &taskCopy, nil
}
