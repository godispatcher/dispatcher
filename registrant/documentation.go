package registrant

import (
	"dispatcher/constants"
	"dispatcher/handling"
	"dispatcher/model"
	"encoding/json"
	"errors"
)

func NewDocumentation(documentairst *model.Documentarist) *Documentation {
	documentation := &Documentation{Documentarist: documentairst}
	return documentation
}

type Documentation struct {
	Documentarist *model.Documentarist
}

func (r *Documentation) DocumentEnforcer() {
	r.TransactionEnforcer(r.Documentarist.Input, r.Documentarist.Output)
}

func (r Documentation) TransactionEnforcer(inputDoc *model.Document, outputDoc *model.Document) {
	outputDoc.Department = inputDoc.Department
	outputDoc.Transaction = inputDoc.Transaction
	var lastResponse interface{} = nil
	transaction, err := r.TransactionMatcher(inputDoc)
	if err != nil {
		outputDoc.Type = constants.DOCUMENT_PARSING_ERROR
		outputDoc.Error = err.Error()
		return
	}

	if inputDoc.Type == constants.DOC_TYPE_PROCEDURE {
		r.TransactionProceduring(transaction, outputDoc)
	} else {
		err = r.ParameterPasser(transaction, inputDoc.Form)
		if err != nil {
			outputDoc.Type = constants.DOC_TYPE_ERROR
			outputDoc.Error = err.Error()
			return
		}

		err = r.DocumentVerification(inputDoc, transaction)

		if err != nil {
			outputDoc.Type = constants.DOC_TYPE_ERROR
			outputDoc.Error = err.Error()
			return
		}

		err = r.ProcessTransact(transaction, outputDoc)
		if err != nil {
			outputDoc.Type = constants.DOC_TYPE_ERROR
			outputDoc.Error = err.Error()
			return
		}

		if outputDoc.Type == constants.DOC_TYPE_RESULT {
			lastResponse = outputDoc.Output
		}

		if lastResponse != nil && inputDoc.ChainRequestOption != nil {
			lastResponse = r.ResponseTransformer(lastResponse, inputDoc.ChainRequestOption)
		}
	}

	if inputDoc.Dispatchings != nil {
		for _, val := range inputDoc.Dispatchings {
			val.Form.FromInterface(lastResponse)
			outputDocNew := &model.Document{}
			if val.Security == nil || val.Security.Licence == "" {
				val.Security = &model.Security{}
				val.Security.Licence = inputDoc.Security.Licence
			}
			r.TransactionEnforcer(val, outputDocNew)
			outputDoc.Dispatchings = append(outputDoc.Dispatchings, outputDocNew)
			if outputDocNew.Type == constants.DOC_TYPE_ERROR {
				return
			}
		}
	}
}

func (r Documentation) TransactionMatcher(inputDoc *model.Document) (transaction *model.Transaction, err error) {
FirstLoop:
	for _, department := range DepartmentRegistering {
		if department.Name == inputDoc.Department {
			for name, findedTransaction := range department.Transactions {
				if name == inputDoc.Transaction {
					transaction = &findedTransaction
					break FirstLoop
				}
			}
		}
	}

	if transaction == nil {
		err = errors.New(constants.TRANSACTION_NOT_FOUND)
	}

	return transaction, err
}

func (r Documentation) ParameterPasser(transaction *model.Transaction, form model.DocumentForm) error {
	formString, _ := formToString(form)
	(*transaction).SetRequest(formString)
	documentFormValidater := handling.DocumentFormValidater{Request: formString}
	reqType := (*transaction).GetRequestType()

	return documentFormValidater.Validate(reqType)
}

func (r Documentation) DocumentVerification(inputDoc *model.Document, transaction *model.Transaction) error {
	if (*transaction).GetOptions().Security.LicenceChecker {
		token := inputDoc.Security.Licence

		if token != "" {
			isValidToken := (*transaction).LicenceChecker(token)
			(*transaction).SetToken(inputDoc.Security.Licence)
			if !isValidToken {
				return errors.New("licence not found")
			}
		} else {
			return errors.New("licence not found")
		}
	}
	return nil
}

func (r Documentation) TransactionProceduring(transaction *model.Transaction, outputDoc *model.Document) {
	inputProcedure := &model.Procedure{}
	outputProcedure := &model.Procedure{}
	inputProcedure.FromRequestType((*transaction).GetRequestType())
	outputProcedure.FromResponseType((*transaction).GetResponse())
	outputDoc.Output = outputProcedure
	transactionOptions := (*transaction).GetOptions()
	outputDoc.Options = &transactionOptions
	outputDoc.Type = constants.DOC_TYPE_PROCEDURE
}

func (r Documentation) ProcessTransact(transaction *model.Transaction, outputDoc *model.Document) error {
	err := (*transaction).Transact()
	outputDoc.Output = (*transaction).GetResponse()
	outputDoc.Type = constants.DOC_TYPE_RESULT

	return err
}

func (r Documentation) ResponseTransformer(response interface{}, chainRequestOption model.ChainRequestOption) (responseMap map[string]interface{}) {
	responseByte, _ := json.Marshal(response)
	json.Unmarshal(responseByte, &responseMap)
	for key, val := range chainRequestOption {
		if _, ok := responseMap[key]; ok {
			responseMap[val.(string)] = responseMap[key]
		}
	}

	return
}
