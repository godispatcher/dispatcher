package registrant

import (
	"dispatcher/constants"
	"dispatcher/handling"
	"dispatcher/model"
	"encoding/json"
	"errors"
	"net/http"
)

func NewRegisterDispatch() RegisterDispatcher {
	dispatch := RegisterDispatcher{}

	dispatch.MainFunc = func(rw http.ResponseWriter, req *http.Request) {
		body, err := handling.RequestHandle(req)
		if err != nil {
			errDoc := model.Document{}
			errResponder := model.New(rw, errDoc)
			errResponder.WriteError(err)
		}
		inputDoc := handling.RequestBodyToDocument(body)
		documentarist := model.NewDocumentarist(rw, inputDoc)
		documentation := NewDocumentation(&documentarist)

		documentation.DocumentEnforcer()
		documentarist.Write()

		/*
			transaction, err := MatchDepartmentAndTransaction(inputDoc)
			documentarist := model.NewDocumentarist(rw, inputDoc)

			if err != nil {
				documentarist.WriteError(err)
				return
			}
			err = transactionRunner(transaction, documentarist.Input, documentarist.Output)

			if err != nil {
				documentarist.WriteError(err)
				return
			}

			var lastResponse interface{} = nil

			if documentarist.Output.Type == constants.DOC_TYPE_RESULT {
				lastResponse = documentarist.Output.Output
			}

			if lastResponse != nil && documentarist.Input.ChainRequestOption != nil {
				lastResponse = responseTransformer(lastResponse, documentarist.Input.ChainRequestOption)
			}

			if documentarist.Input.Dispatchings != nil {
				for _, val := range documentarist.Input.Dispatchings {
					val.Form.FromInterface(lastResponse)
					outputDocNew := &model.Document{}
					if &val.Security.Licence == nil || val.Security.Licence == "" {
						val.Security.Licence = documentarist.Input.Security.Licence
					}
					err := dispatchTracker(val, outputDocNew)
					if err != nil {
						outputDocNew.Error = err.Error()
						outputDocNew.Type = constants.DOC_TYPE_ERROR
					}
					documentarist.Output.Dispatchings = append(documentarist.Output.Dispatchings, outputDocNew)
				}
			}
			documentarist.WriteDocument()
		*/
	}

	dispatch.Port = "9000"
	return dispatch
}

func DocumentHandler(inputDoc *model.Document, outputDoc *model.Document) {
	transaction, err := MatchDepartmentAndTransaction(inputDoc)
	if err != nil {
		outputDoc.Error = err.Error()
		return
	}
	err = transactionRunner(transaction, inputDoc, outputDoc)

	if err != nil {
		outputDoc.Error = err.Error()
		return
	}

	var lastResponse interface{} = nil

	if outputDoc.Type == constants.DOC_TYPE_RESULT {
		lastResponse = outputDoc.Output
	}

	if lastResponse != nil && inputDoc.ChainRequestOption != nil {
		lastResponse = responseTransformer(lastResponse, inputDoc.ChainRequestOption)
	}

	if inputDoc.Dispatchings != nil {
		for _, val := range inputDoc.Dispatchings {
			val.Form.FromInterface(lastResponse)
			outputDocNew := &model.Document{}
			err := dispatchTracker(val, outputDocNew)
			if err != nil {
				outputDocNew.Error = err.Error()
				outputDocNew.Type = constants.DOC_TYPE_ERROR
			}
			outputDoc.Dispatchings = append(outputDoc.Dispatchings, outputDocNew)
		}
	}

	return
}

func dispatchTracker(inputDoc *model.Document, outputDoc *model.Document) (err error) {
	transaction, err := MatchDepartmentAndTransaction(inputDoc)
	if err != nil {
		return err
	}
	err = transactionRunner(transaction, inputDoc, outputDoc)
	if err != nil {
		return err
	}
	var lastResponse interface{} = nil
	if outputDoc.Type == constants.DOC_TYPE_RESULT {
		lastResponse = outputDoc.Output
	}

	if lastResponse != nil && inputDoc.ChainRequestOption != nil {
		lastResponse = responseTransformer(lastResponse, inputDoc.ChainRequestOption)
	}
	// TODO: lastResponse doc.ChainRequestOption a göre işlemden geçirildikten sonra val.Form.FromInterface fonksiyonuna aktarılacak
	// Ana blokta da bu işlemin aynı gerekiyor.
	if inputDoc.Dispatchings != nil {
		for _, val := range inputDoc.Dispatchings {
			val.Form.FromInterface(lastResponse)
			outputDocNew := &model.Document{}
			if &val.Security.Licence == nil || val.Security.Licence == "" {
				val.Security.Licence = inputDoc.Security.Licence
			}
			err = dispatchTracker(val, outputDocNew)
			if err != nil {
				return err
			}
		}
	}

	return
}

func transactionRunner(transaction *model.Transaction, inputDoc *model.Document, outputDoc *model.Document) (err error) {
	if inputDoc.Type == constants.DOC_TYPE_PROCEDURE {
		inputProcedure := &model.Procedure{}
		outputProcedure := &model.Procedure{}
		inputProcedure.FromRequestType((*transaction).GetRequestType())
		outputProcedure.FromResponseType((*transaction).GetResponse())
		outputDoc.Output = outputProcedure
		transactionOptions := (*transaction).GetOptions()
		outputDoc.Options = &transactionOptions
		outputDoc.Type = constants.DOC_TYPE_PROCEDURE
		return nil
	}
	err = RequestHandler(inputDoc, transaction)
	if err != nil {
		return err
	}

	if transaction != nil {
		if (*transaction).GetOptions().Security.LicenceChecker {
			token := inputDoc.Security.Licence

			if token != "" {
				isValidToken := (*transaction).LicenceChecker(token)
				if !isValidToken {
					return errors.New("licence not found")
				}
			} else {
				return errors.New("licence not found")
			}
		}
		err := (*transaction).Transact()
		outputDoc.Output = (*transaction).GetResponse()
		outputDoc.Type = constants.DOC_TYPE_RESULT
		return err
	}

	return errors.New("an unidentified error has occurred")
}

type RegisterDispatcher struct {
	MainFunc func(http.ResponseWriter, *http.Request)
	Port     string
}

func responseTransformer(response interface{}, chainRequestOption model.ChainRequestOption) (responseMap map[string]interface{}) {
	responseByte, _ := json.Marshal(response)
	json.Unmarshal(responseByte, &responseMap)
	for key, val := range chainRequestOption {
		if _, ok := responseMap[key]; ok {
			responseMap[val.(string)] = responseMap[key]
		}
	}

	return
}
