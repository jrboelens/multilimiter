package multilimiter

import "errors"

var LimiterStopped = errors.New("Limiter has been stopped")
var DeadlineExceeded = errors.New("Timeout Exceeded")
