package handling

import (
	"github.com/denizakturk/dispatcher/constants"
	"github.com/denizakturk/dispatcher/model"
	"encoding/json"
	"log"
)

func RequestBodyToDocument(body []byte) *model.Document {
	document := &model.Document{}
	err := json.Unmarshal(body, &document)

	if err != nil {
		log.Printf(constants.DOCUMENT_PARSING_ERROR, err)
	}

	return document
}

type TransactionExchangeConverter struct{}

func (c TransactionExchangeConverter) HandleRequest(req string, reqType interface{}) error {
	return json.Unmarshal([]byte(req), &reqType)
}

func (c TransactionExchangeConverter) HandleResponse(responseType interface{}) (string, error) {
	responseByte, err := json.Marshal(responseType)

	return string(responseByte), err
}
