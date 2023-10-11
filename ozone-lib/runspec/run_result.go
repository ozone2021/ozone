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

func (r *RunResult) findCallstackResult(name string) (*CallstackResult, error) {
	for callstackResult := range r.CallstackResults {
		if callstackResult.Name == name {
			return callstackResult, nil
		}
	}
	return nil, errors.New("CallstackResult not found")
}

func (r *RunResult) AddRootCallstack(callstack *CallStack) {
	r.CallstackResults[NewSucceededCallstackResult(callstack.RootRunnableName)] = []*CallstackResult{}
}

func (r *RunResult) AddCallstackResult(rootRunnableName string, callstackResults []*CallstackResult, err error) error {
	if len(callstackResults) == 0 {
		return errors.New("CallstackResults must have at least one result")
	}
	subCallstackResults := callstackResults[0:]

	rootCallstackResult, err := r.findCallstackResult(rootRunnableName)
	r.CallstackResults[rootCallstackResult] = subCallstackResults

	for _, result := range callstackResults {
		if result.Status == Failed {
			r.Status = Failed
			return nil
		}
	}

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

func NewFailedCallstackResult(name string, err error) *CallstackResult {
	return &CallstackResult{
		Status: Failed,
		Name:   name,
		Err:    err,
	}
}

func NewCachedCallstackResult(name string) *CallstackResult {
	return &CallstackResult{
		Status: Cached,
		Name:   name,
	}
}
