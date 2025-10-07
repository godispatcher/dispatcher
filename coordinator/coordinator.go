package coordinator

import (
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

type ServiceRequest struct {
	Host     string
	Document model.Document
}

func CallTransaction(request ServiceRequest) (model.Document, error) {
	return server.CallHTTP(request.Host, request.Document)
}
