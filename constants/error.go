package constants

const (
	FIELD_NOT_FOUND                string = "the field named %s is requre but not found"
	FIELD_CANNOT_BE_EMPTY          string = "the field named %s is marked as cannot be empty"
	DOCUMENT_PARSING_ERROR         string = "error document parsing %v"
	DOCUMENT_VERIFICATION_FAILED   string = "document verification failed"
	TRANSACTION_NOT_FOUND          string = "transaction is not found"
	REQUEST_BODY_READ_ERROR        string = "error reading body: %v"
	THIS_REQUEST_TYPE_INVALID_JSON string = "this request data is broken"
	CONTENT_TYPE_NOT_JSON          string = "in this request, header content type is not marked as json, add content-type:application/json to request header to fix it."
)
