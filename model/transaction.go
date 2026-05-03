package model

type Transaction interface {
	Transact() (interface{}, error)
	GetRequest() interface{}
	GetResponse() interface{}
}

type SecurityOptions struct {
	LicenceChecker bool `json:"licence_checker,omitempty"`
}

type RateLimiterScope string

const (
	ScopeGlobal RateLimiterScope = "global"
	ScopeIP     RateLimiterScope = "ip"
	ScopeUser   RateLimiterScope = "user"
	ScopeApiKey RateLimiterScope = "api_key"
	ScopeRoute  RateLimiterScope = "route"
)

type RateLimitOptions struct {
	Enabled bool             `json:"enabled,omitempty" yaml:"enabled"`
	Limit   int              `json:"limit,omitempty" yaml:"limit"`
	Window  int              `json:"window,omitempty" yaml:"window"` // seconds
	Scope   RateLimiterScope `json:"scope,omitempty" yaml:"scope"`
}

type TransactionOptions struct {
	Security    SecurityOptions  `json:"security,omitempty" yaml:"security"`
	RateLimiter RateLimitOptions `json:"rate_limiter,omitempty" yaml:"rate_limiter"`
}

func (m TransactionOptions) GetOptions() TransactionOptions {
	return m
}

type LicenceValidator func(licence string) (isValid bool)
