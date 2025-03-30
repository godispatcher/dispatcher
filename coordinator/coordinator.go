package coordinator

import (
	"github.com/godispatcher/dispatcher/department"
	"github.com/godispatcher/dispatcher/model"
)

func ExecuteTransaction(document model.Document) model.Document {
	transaction := department.DispatcherHolder.GetTransaction(document.Department, document.Transaction)
	if transaction != nil {
		return (*transaction).GetTransaction().Init(document)
	}
	return document
}
