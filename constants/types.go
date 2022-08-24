package constants

const (
	HTTP_CONTENT_TYPE = "Content-Type"

	DOC_TYPE_RESULT          = "Result"          // Transaction Result
	DOC_TYPE_ERROR           = "Error"           // Error
	DOC_TYPE_PROCEDURE       = "Procedure"       // Transactiın procedure parameters
	DOC_TYPE_DISPATCH        = "Dispatch"        // Dispatch to transaction and/or fill a form
	DOC_TYPE_DIRECT_DISPATCH = "Direct Dispatch" // Direct Dispatch to transaction no form filling (Require form transactions)
)
