package model

type Transaction interface {
	Transact() (interface{}, error)
	GetRequest() interface{}
	GetResponse() interface{}
}

type TransactionOptions struct {
	Security    SecurityOptions    `json:"security,omitempty"`
	RateLimiter RateLimiterOptions `json:"rate_limiter,omitempty"`
}

func (m TransactionOptions) GetOptions() TransactionOptions {
	return m
}

type SecurityOptions struct {
	LicenceChecker bool `json:"licence_checker,omitempty"`
}

type RateLimiterOptions struct {
	Enabled bool `json:"enabled,omitempty"`
	Limit   int  `json:"limit,omitempty"`
	Window  int  `json:"window,omitempty"` // seconds
}

type LicenceValidator func(licence string) (isValid bool)
