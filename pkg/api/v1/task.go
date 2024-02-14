package api

import (
	"encoding/json"
	"net/http"

	"github.com/deividaspetraitis/ledger"

	"github.com/gorilla/mux"
)

// NewCreateTaskResponse creates and returns a new CreateTaskResponse.
func NewCreateTaskResponse(id string) *CreateTaskResponse {
	return &CreateTaskResponse{
		ID: id,
	}
}

// CreateTaskResponse represents HTTP response for creating a new task.
type CreateTaskResponse struct {
	ID string `json:"id"` // Task ID
}

// MarshalHTTP implements http.Marshaler.
func (r *CreateTaskResponse) MarshalHTTP(w http.ResponseWriter) error {
	return json.NewEncoder(w).Encode(r)
}

// Task represents API response Task entity.
type Task struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// NewTaskResponse constructs and returns response Task entity.
func NewTaskResponse(t *ledger.Task) *Task {
	return &Task{
		ID:     t.ID,
		Status: t.Status,
	}
}

// MarshalHTTP implements http.Marshaler.
func (r *Task) MarshalHTTP(w http.ResponseWriter) error {
	return json.NewEncoder(w).Encode(r)
}

// GetTaskRequest represents HTTP request for retrieving a task.
type GetTaskRequest struct {
	ID string `json:"id"` // Task ID
}

// Validate validates request data and returns an error if it's not a valid.
// Validate implements validator.Validator.
// TODO: implement more sophisticated rule
func (r *GetTaskRequest) Validate() error {
	if len(r.ID) < 3 {
		return ledger.ErrNotValidTaskID
	}
	return nil
}

// UnmarshalHTTP implements http.RequestUnmarshaler.
func (r *GetTaskRequest) UnmarshalHTTPRequest(req *http.Request) error {
	r.ID = mux.Vars(req)["id"]
	return r.Validate()
}

// Parse parses and returns Task ID from the request.
func (r *GetTaskRequest) Parse() string {
	return r.ID
}
