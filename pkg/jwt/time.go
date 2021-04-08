package jwt

import (
	"encoding/json"
	"time"
)

// Time wraps time.Time for usage in JWT claims.
// It implements proper JSON marshaling
// according to NumericDate in https://tools.ietf.org/html/rfc7519#section-2.
type Time time.Time

// MarshalJSON marshals the given time to unix epoch time
func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Unix())
}

// UnmarshalJSON parses the given src as unix epoch time
func (t *Time) UnmarshalJSON(src []byte) error {
	var sec int64
	if err := json.Unmarshal(src, &sec); err != nil {
		return err
	}

	*t = Time(time.Unix(sec, 0))
	return nil
}
