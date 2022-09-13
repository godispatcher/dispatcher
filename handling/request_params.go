package handling

import (
	"encoding/json"
	"fmt"

	"github.com/denizakturk/dispatcher/constants"
	"github.com/denizakturk/dispatcher/model"
)

func RequestBodyToDocument(body []byte) (*model.Document, error) {
	document := &model.Document{}
	err := json.Unmarshal(body, &document)

	if err != nil {
		return nil, fmt.Errorf(constants.DOCUMENT_PARSING_ERROR, err)
	}

	return document, nil
}

type TransactionExchangeConverter struct{}

func (c TransactionExchangeConverter) HandleRequest(req string, reqType interface{}) error {
	return json.Unmarshal([]byte(req), &reqType)
}

func (c TransactionExchangeConverter) HandleResponse(responseType interface{}) (string, error) {
	responseByte, err := json.Marshal(responseType)

	return string(responseByte), err
}
