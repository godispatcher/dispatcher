package coordinator

import (
	"encoding/json"

	"github.com/godispatcher/dispatcher/department"
	"github.com/godispatcher/dispatcher/model"
	"github.com/godispatcher/dispatcher/server"
)

func ExecuteTransaction(document model.Document) model.Document {
	transaction := department.DispatcherHolder.GetTransaction(document.Department, document.Transaction)
	if transaction != nil {
		return (*transaction).GetTransaction().Init(document)
	}
	return document
}

// ServiceRequest is a generic request wrapper for calling remote transactions
// T is the request form model, R is the expected response output model
// Uses model.Document directly; callers should populate non-form fields on Document.
// Note: host should be provided without protocol (e.g., "auth:9000").
// server.CallHTTP will construct the full URL.

type ServiceRequest[T any, R any] struct {
	Address  string
	Document model.Document
	Request  T
}

// CallTransaction sends the typed request T and returns a typed response R.
// Internally it fills Document.Form from T, calls the HTTP client and decodes Output into R.
func (req ServiceRequest[T, R]) CallTransaction() (R, error) {
	var zero R
	// Build form from typed request into the provided document
	form := model.DocumentForm{}
	if err := form.FromInterface(req.Request); err != nil {
		return zero, err
	}
	req.Document.Form = form
	// Ensure verify code is propagated to the outgoing request document
	if (req.Document.Security == nil) || (req.Document.Security.VerifyCode == "") {
		if vc := model.GetCurrentVerifyCode(); vc != "" {
			if req.Document.Security == nil {
				req.Document.Security = &model.Security{}
			}
			req.Document.Security.VerifyCode = vc
		}
	}
	resDoc, err := server.CallHTTP(req.Address, req.Document)
	if err != nil {
		return zero, err
	}
	// Decode Output into typed response
	b, err := json.Marshal(resDoc.Output)
	if err != nil {
		return zero, err
	}
	var out R
	if err := json.Unmarshal(b, &out); err != nil {
		return zero, err
	}
	return out, nil
}
