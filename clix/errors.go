package clix

import (
	"errors"
)

var (
	ErrorTypeNotSupported = errors.New("Type not supported")

	errorNotStarted = errors.New("Not started")
)

type AggregatedError struct {
	Errors []error
}

func (e *AggregatedError) AddErr(err error) error {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
	return err
}

func (e *AggregatedError) Add(err error) bool {
	return e.AddErr(err) != nil
}

func (e *AggregatedError) AddMany(errs ...error) *AggregatedError {
	for _, err := range errs {
		e.AddErr(err)
	}
	return e
}

func (e *AggregatedError) Aggregate() error {
	if len(e.Errors) > 0 {
		return e
	}
	return nil
}

func (e *AggregatedError) Error() string {
	if len(e.Errors) > 0 {
		msg := "Multiple Errors:"
		for _, err := range e.Errors {
			msg += "\n" + err.Error()
		}
		return msg
	}
	return ""
}
