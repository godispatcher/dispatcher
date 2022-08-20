package converter

import (
	"encoding/json"
	"log"
	"dispatcher/constants"
	"dispatcher/document"
)

func RequestBodyToDocument(body []byte) *document.Document {
	document := &document.Document{}
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
