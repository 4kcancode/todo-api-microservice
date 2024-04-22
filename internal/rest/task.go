package rest

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
=======
	"github.com/go-chi/chi/v5"

	"github.com/MarioCarrion/todo-api/internal"
)

//go:generate counterfeiter -generate

//counterfeiter:generate -o resttesting/task_service.gen.go . TaskService

// TaskService ...
type TaskService interface {
	By(ctx context.Context, args internal.SearchParams) (internal.SearchResults, error)
	Create(ctx context.Context, params internal.CreateParams) (internal.Task, error)
	Delete(ctx context.Context, id string) error
	Task(ctx context.Context, id string) (internal.Task, error)
	Update(ctx context.Context, id string, description string, priority internal.Priority, dates internal.Dates, isDone bool) error
}

// TaskHandler ...
type TaskHandler struct {
	svc TaskService
}

// NewTaskHandler ...
func NewTaskHandler(svc TaskService) *TaskHandler {
	return &TaskHandler{
		svc: svc,
	}
}

// Register connects the handlers to the router.
func (t *TaskHandler) Register(r *echo.Echo) {
	r.POST("/tasks", t.create)
	r.GET("/tasks/:id", t.task)
	r.PUT("/tasks/:id", t.update)
	r.DELETE("/tasks/:id", t.delete)
	r.POST("/search/tasks", t.search)
=======
func (t *TaskHandler) Register(r *chi.Mux) {
	r.Post("/tasks", t.create)
	r.Get(fmt.Sprintf("/tasks/{id:%s}", uuidRegEx), t.task)
	r.Put(fmt.Sprintf("/tasks/{id:%s}", uuidRegEx), t.update)
	r.Delete(fmt.Sprintf("/tasks/{id:%s}", uuidRegEx), t.delete)
  r.Post("/search/tasks", t.search)
}

// Task is an activity that needs to be completed within a period of time.
//
//nolint:tagliatelle
type Task struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Priority    Priority `json:"priority"`
	Dates       Dates    `json:"dates"`
	IsDone      bool     `json:"is_done"`
}

// CreateTasksRequest defines the request used for creating tasks.
type CreateTasksRequest struct {
	Description string   `json:"description"`
	Priority    Priority `json:"priority"`
	Dates       Dates    `json:"dates"`
}

// CreateTasksResponse defines the response returned back after creating tasks.
type CreateTasksResponse struct {
	Task Task `json:"task"`
}

func (t *TaskHandler) create(router echo.Context) error {
	var req CreateTasksRequest
	if err := router.Bind(&req); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "json decoder")
=======
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderErrorResponse(w, r, "invalid request",
			internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "json decoder"))

		return
	}

	task, err := t.svc.Create(router.Request().Context(), internal.CreateParams{
		Description: req.Description,
		Priority:    req.Priority.Convert(),
		Dates:       req.Dates.Convert(),
	})
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "create failed")
	}

	return router.JSON(http.StatusCreated,
=======
		renderErrorResponse(w, r, "create failed", err)

		return
	}

	renderResponse(w, r,
		&CreateTasksResponse{
			Task: Task{
				ID:          task.ID,
				Description: task.Description,
				Priority:    NewPriority(task.Priority),
				Dates:       NewDates(task.Dates),
			},
		})
}

func (t *TaskHandler) delete(router echo.Context) error {
	id := router.Param("id")
=======
func (t *TaskHandler) delete(w http.ResponseWriter, r *http.Request) {
	// NOTE: Safe to ignore error, because it's always defined.
	id := chi.URLParam(r, "id")

	if err := t.svc.Delete(r.Context(), id); err != nil {
		renderErrorResponse(w, r, "delete failed", err)

	if err := t.svc.Delete(router.Request().Context(), id); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "delete failed")
	}

	return router.JSON(http.StatusOK, struct{}{})
=======
	renderResponse(w, r, struct{}{}, http.StatusOK)
}

// ReadTasksResponse defines the response returned back after searching one task.
type ReadTasksResponse struct {
	Task Task `json:"task"`
}

func (t *TaskHandler) task(router echo.Context) error {
	id := router.Param("id")
=======
func (t *TaskHandler) task(w http.ResponseWriter, r *http.Request) {
	// NOTE: Safe to ignore error, because it's always defined.
	id := chi.URLParam(r, "id")
	task, err := t.svc.Task(router.Request().Context(), id)
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "find failed")
	}

	return router.JSON(http.StatusOK,
=======
		renderErrorResponse(w, r, "find failed", err)

		return
	}

	renderResponse(w, r,
		&ReadTasksResponse{
			Task: Task{
				ID:          task.ID,
				Description: task.Description,
				Priority:    NewPriority(task.Priority),
				Dates:       NewDates(task.Dates),
				IsDone:      task.IsDone,
			},
		})
}

// UpdateTasksRequest defines the request used for updating a task.
//
//nolint:tagliatelle
type UpdateTasksRequest struct {
	Description string   `json:"description"`
	IsDone      bool     `json:"is_done"`
	Priority    Priority `json:"priority"`
	Dates       Dates    `json:"dates"`
}

func (t *TaskHandler) update(router echo.Context) error {
	var req UpdateTasksRequest
	if err := router.Bind(&req); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "json decoder")
	}

	id := router.Param("id")
=======
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderErrorResponse(w, r, "invalid request",
			internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "json decoder"))

		return
	}

	defer r.Body.Close()

	// NOTE: Safe to ignore error, because it's always defined.
	id := chi.URLParam(r, "id")

	err := t.svc.Update(router.Request().Context(), id, req.Description, req.Priority.Convert(), req.Dates.Convert(), req.IsDone)
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "update failed")
	}

	return router.JSON(http.StatusOK, struct{}{})
=======
		renderErrorResponse(w, r, "update failed", err)

		return
	}

	renderResponse(w, r, &struct{}{}, http.StatusOK)
}

// SearchTasksRequest defines the request used for searching tasks.
//
//nolint:tagliatelle
type SearchTasksRequest struct {
	Description *string   `json:"description"`
	Priority    *Priority `json:"priority"`
	IsDone      *bool     `json:"is_done"`
	From        int64     `json:"from"`
	Size        int64     `json:"size"`
}

// SearchTasksResponse defines the response returned back after searching for any task.
type SearchTasksResponse struct {
	Tasks []Task `json:"tasks"`
	Total int64  `json:"total"`
}

func (t *TaskHandler) search(router echo.Context) error {
	var req SearchTasksRequest
	if err := router.Bind(&req); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "json decoder")
=======
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderErrorResponse(w, r, "invalid request",
			internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "json decoder"))

		return
	}

	var priority *internal.Priority

	if req.Priority != nil {
		res := req.Priority.Convert()
		priority = &res
	}

	res, err := t.svc.By(router.Request().Context(), internal.SearchParams{
		Description: req.Description,
		Priority:    priority,
		IsDone:      req.IsDone,
		From:        req.From,
		Size:        req.Size,
	})
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "search failed")
=======
		renderErrorResponse(w, r, "search failed", err)

		return
	}

	tasks := make([]Task, len(res.Tasks))

	for i, task := range res.Tasks {
		tasks[i].ID = task.ID
		tasks[i].Description = task.Description
		tasks[i].Priority = NewPriority(task.Priority)
		tasks[i].Dates = NewDates(task.Dates)
	}

	return router.JSON(http.StatusOK,
=======
	renderResponse(w, r,
		&SearchTasksResponse{
			Tasks: tasks,
			Total: res.Total,
		})
}
