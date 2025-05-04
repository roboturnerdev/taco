package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
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

// make a custom type for a "test case"
type creatWorkstreamTestCase struct {
	name				string
	formData			url.Values
	mockCreateFunc		func(ws models.Workstream) error
	expectedStatusCode	int
	expectedRedirect	bool
	expectedWorkstream	models.Workstream
}

// so that i may make a list (slice, array) of test case objects
var createWorkstreamTestCases = []creatWorkstreamTestCase{
	{
		name: "Valid workstream",
		formData: url.Values{
			"name": 		{"Test Workstream"},
			"code": 		{"ABC123"},
			"location": 	{"New York"},
			"description": 	{"A test workstream"},
			"identity": 	{"DevOps"},
			"quote": 		{"Ship it!"},
		},
		mockCreateFunc: func(ws models.Workstream) error {
			return nil // fake success
		},
		expectedStatusCode: http.StatusFound,
		expectedRedirect: true,
		expectedWorkstream: models.Workstream{
			Name: 			"Test Workstream",
			Code: 			"ABC123",
			Location: 		"New York",
			Description: 	"A test workstream",
			Identity: 		"DevOps",
			Quote: 			"Ship it!",
		},
	},
	{
		name: "Error creating workstream",
		formData: url.Values{
			"name":        {"Test Workstream"},
			"code":        {"ABC123"},
			"location":    {"New York"},
			"description": {"A test workstream"},
			"identity":    {"DevOps"},
			"quote":       {"Ship it!"},
		},
		mockCreateFunc: func(ws models.Workstream) error {
			return fmt.Errorf("error creating workstream") // fake error
		},
		expectedStatusCode: http.StatusInternalServerError,
		expectedRedirect:   false,
		expectedWorkstream: models.Workstream{
			Name: 			"Test Workstream",
			Code: 			"ABC123",
			Location: 		"New York",
			Description: 	"A test workstream",
			Identity: 		"DevOps",
			Quote: 			"Ship it!",
		},
	},
}

func TestCreateNewPost(t *testing.T){
	
	// one logger for all handler testing
	logger := log.New(io.Discard, "", 0)

	for _, tc := range createWorkstreamTestCases {
		t.Run(tc.name, func(t *testing.T) {

			// Arrange
				
			// - make a dummy db
			called := false
			var savedWorkstream models.Workstream
			mockDB := &mockWorkstreamDB{
				CreateWorkstreamFunc: func(ws models.Workstream) error {
					called = true
					savedWorkstream = ws
					return tc.mockCreateFunc(ws)
				},
			}

			// mock the handler with dummy db
			h := handler.NewWorkstreamHandler(logger, mockDB)
			
			// Act

			// build the request
			req := httptest.NewRequest(http.MethodPost, "/workstreams/new", strings.NewReader(tc.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			// send request and capture response
			h.CreateNewPost(w, req)
			res := w.Result()

			// Assert

			// case: expected status code
			// failure: logic failed to provide expected status code
			if res.StatusCode != tc.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", tc.expectedStatusCode, res.StatusCode)
			}
			// case: expected to be redirected after creation
			// failure: creating new workstream did not redirect
			if tc.expectedRedirect && res.StatusCode != http.StatusFound {
				t.Errorf("expected redirect (302), got %d", res.StatusCode)
			}
			// case: expected to fail redirect
			// failure: logic should have stopped me from being redirected
			if !tc.expectedRedirect && res.StatusCode == http.StatusFound {
				t.Errorf("expected no redirect, got %d", res.StatusCode)
			}
			// case: expected saved workstream to match
			// failure: new workstream is not being saved correctly
			if called && !reflect.DeepEqual(savedWorkstream, tc.expectedWorkstream) {
				t.Errorf("expected workstream %+v, got %+v", tc.expectedWorkstream, savedWorkstream)
			}
			// case: CreateWorkstream was not called
			// failure: create workstream handler didn't call the expected function
			if !called && tc.mockCreateFunc != nil {
				t.Errorf("CreateWorkstream was not called.")
			}
			// case: resulting new workstream name didnt match expected
			// failure: new workstream names are being saved wrong
			if savedWorkstream.Name != "Test Workstream" {
				t.Errorf("expected name to be 'Test Workstream', got '%s'", savedWorkstream.Name)
			}
		})
	}
}
