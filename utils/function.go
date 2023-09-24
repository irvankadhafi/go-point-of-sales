package utils

import (
	"github.com/sirupsen/logrus"
	"time"
)

type RetryStopper struct {
	error
}

func Retry(attempts int, sleep time.Duration, cbFn func() error) error {
	if err := cbFn(); err != nil {
		if s, ok := err.(RetryStopper); ok {
			return s.error
		}

		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return Retry(attempts, 2*sleep, cbFn)
		}
		return err
	}
	return nil
}

// WrapCloser call close and log the error
func WrapCloser(close func() error) {
	if close == nil {
		return
	}
	if err := close(); err != nil {
		logrus.Error(err)
	}
}
