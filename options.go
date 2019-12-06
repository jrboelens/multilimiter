package multilimiter

const DEFAULT_RATE = 1.0
const DEFAULT_CONCURRENCY = 1

// Base interface for all options
type Option interface {
	apply(*options)
}

// Contains all possible options
type options struct {
	rateLimit *RateLimitOption
	concLimit *ConcLimitOption
}

// Creates an instance of options out of a slice of Options
func CreateOptions(opts ...Option) *options {
	allOpts := &options{}

	for _, opt := range opts {
		opt.apply(allOpts)
	}

	setDefaultOpts(allOpts)
	return allOpts
}

func setDefaultOpts(allOpts *options) {
	if allOpts.rateLimit == nil {
		allOpts.rateLimit = &RateLimitOption{NewRateLimiter(DEFAULT_RATE)}
	}
	if allOpts.concLimit == nil {
		allOpts.concLimit = &ConcLimitOption{NewConcLimiter(DEFAULT_CONCURRENCY)}
	}
}

// option for controlling rate limiting
type RateLimitOption struct {
	Limiter RateLimiter
}

func (me *RateLimitOption) apply(allopts *options) {
	allopts.rateLimit = me
}

// option for controlling concurrency
type ConcLimitOption struct {
	Limiter ConcLimiter
}

func (me *ConcLimitOption) apply(allopts *options) {
	allopts.concLimit = me
}
