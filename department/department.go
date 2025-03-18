package department

import (
	"github.com/godispatcher/dispatcher/model"
	"github.com/godispatcher/dispatcher/transaction"
	"github.com/godispatcher/logger"
	"net/http"
	"time"
)

type Department struct {
	Name         string
	Transactions []*transaction.TransactionBucketItemInterface
}

type DispacherBucket []*Department

func (db *DispacherBucket) Add(name string, transaction transaction.TransactionBucketItemInterface) {
	hasDepartment := false
	hasTransaction := false
	for _, val := range *db {
		if val.Name == name {
			hasDepartment = true
			val.Transactions = append(val.Transactions, &transaction)
		}
	}
	if !hasDepartment && !hasTransaction {
		tmpDep := &Department{}
		tmpDep.Name = name
		tmpDep.Transactions = append(tmpDep.Transactions, &transaction)
		*db = append(*db, tmpDep)
	}
}

func (db *DispacherBucket) GetTransaction(departmentName, transactionName string) *transaction.TransactionBucketItemInterface {
	for _, val := range *db {
		if val.Name == departmentName {
			for _, v := range val.Transactions {
				for (*v).GetName() == transactionName {
					return v
				}
			}
		}
	}

	return nil
}

var DispatcherHolder DispacherBucket

func NewRegisteryDispatcher(port string) *RegisterDispatcher {
	return &RegisterDispatcher{Port: port, MainFunc: RegisterMainFunc}
}

type RegisterDispatcher struct {
	MainFunc func(http.ResponseWriter, *http.Request) model.RegisterResponseModel
	Port     string
}

func (rd RegisterDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.InitLogFile("log.jsonl")
	loggerRequest, _ := logger.NewLoggedRequest(r)
	startTime := time.Now()
	rw := rd.MainFunc(w, r)
	duration := time.Since(startTime)
	loggerResponse := logger.NewLoggedResponse(rw.StatusCode, rw.Header, rw.Body)
	entry := logger.LogEntry{
		Timestamp: time.Now(),
		Request:   loggerRequest,
		Response:  loggerResponse,
		Duration:  duration,
	}
	logger.WriteLog(entry)
}
