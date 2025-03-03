package exoclick

import (
	"time"
)

type Rate struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

func (r Rate) String() string {
	return Stringify(r)
}
