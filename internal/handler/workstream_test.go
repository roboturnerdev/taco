package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"taco/internal/handler"
	"taco/internal/models"
	"testing"

	"io"
	"log"
)

// 1. Mock the DB to satisfy the interface (dummy db)
type mockWorkstreamDB struct {
	CreateWorkstreamFunc 	func(workstream models.Workstream) error
	GetAllWorkstreamsFunc   func() ([]models.Workstream, error)
	GetWorkstreamByIDFunc   func(id int) (models.Workstream, error)
	DeleteWorkstreamFunc    func(id int) error
}

func (m *mockWorkstreamDB) CreateWorkstream(ws models.Workstream) error {
	if m.CreateWorkstreamFunc != nil {
		return m.CreateWorkstreamFunc(ws)
	}
	return nil
}

func (m *mockWorkstreamDB) GetAllWorkstreams() ([]models.Workstream, error) {
	if m.GetAllWorkstreamsFunc != nil {
		return m.GetAllWorkstreamsFunc()
	}
	return nil, nil
}

func (m *mockWorkstreamDB) GetWorkstreamByID(id int) (models.Workstream, error) {
	if m.GetWorkstreamByIDFunc != nil {
		return m.GetWorkstreamByIDFunc(id)
	}
	return models.Workstream{}, nil
}

func (m *mockWorkstreamDB) DeleteWorkstream(id int) error {
	if m.DeleteWorkstreamFunc != nil {
		return m.DeleteWorkstreamFunc(id)
	}
	return nil
}

func TestCreateNewPost_Success(t *testing.T) {

	// Arrange
	
	// - make a dummy db
	called := false
	var savedWorkstream models.Workstream
	mockDB := &mockWorkstreamDB{
		CreateWorkstreamFunc: func(ws models.Workstream) error {
			called = true
			savedWorkstream = ws
			return nil
		},
	}

	// set up dependencies to implement a mock handler
	logger := log.New(io.Discard, "", 0)

	// mock the handler with dummy logger and db
	h := handler.NewWorkstreamHandler(logger, mockDB)

	// create a body to send to the mock post endpoint handler
	form := url.Values{}
	form.Add("name", "Test Workstream")
	form.Add("code", "ABC123")
	form.Add("location", "New York")
	form.Add("description", "A test workstream")
	form.Add("identity", "DevOps")
	form.Add("quote", "Ship it!")

	// Act
	
	// build the request
	req := httptest.NewRequest(http.MethodPost, "/workstreams/new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// send request and capture response
	h.CreateNewPost(w, req)
	res := w.Result()

	// Assert

	// case: did not get redirected
	if res.StatusCode != http.StatusFound {
		t.Errorf("expected redirect (302), got %d", res.StatusCode)
	}
	// case: CreateWorkstream did not get called
	if !called {
		t.Fatal("expected CreateWorkstream to be called")
	}
	// case: resulting new workstream name didnt match expected 
	if savedWorkstream.Name != "Test Workstream" {
		t.Errorf("expected name to be 'Test Workstream', got '%s'", savedWorkstream.Name)
	}
}