package runspec

import (
	"errors"
	"log"
)

type RunResult struct {
	Status           CallstackStatus
	CallstackResults map[*CallstackResult][]*CallstackResult
}

type CallstackStatus int

const (
	Succeeded = iota
	Failed
	Cached
)

type CallstackResult struct {
	Status CallstackStatus
	Name   string
	Err    error
}

func NewRunResult() *RunResult {
	return &RunResult{
		Status:           Succeeded,
		CallstackResults: make(map[*CallstackResult][]*CallstackResult),
	}
}

func (r *RunResult) AddCallstackResult(callstackResults []*CallstackResult, err error) error {
	if len(callstackResults) == 0 {
		return errors.New("CallstackResults must have at least one result")
	}
	rootCallstack := callstackResults[0]
	rootCallstack.Err = err
	if len(callstackResults) == 1 {
		r.CallstackResults[rootCallstack] = []*CallstackResult{callstackResults[0]}
		return nil
	}
	subCallstackResults := callstackResults[1:]
	r.CallstackResults[rootCallstack] = subCallstackResults

	return nil
}

func (r *RunResult) PrintRunResult() {
	for rootCallstack, subCallstacks := range r.CallstackResults {
		log.Printf("Callstack: %s, Status: %d, Error: %s \n", rootCallstack.Name, rootCallstack.Status, rootCallstack.Err)
		for _, subCallstack := range subCallstacks {
			log.Printf("  Subcallstack: %s, Status: %d, Error: %s \n", subCallstack.Name, subCallstack.Status, subCallstack.Err)
		}
	}
}

func NewSucceededCallstackResult(name string) *CallstackResult {
	return &CallstackResult{
		Status: Succeeded,
		Name:   name,
	}
}

func NewFailedCallstackResult(name string) *CallstackResult {
	return &CallstackResult{
		Status: Failed,
		Name:   name,
	}
}

func NewCachedCallstackResult(name string) *CallstackResult {
	return &CallstackResult{
		Status: Cached,
		Name:   name,
	}
}
