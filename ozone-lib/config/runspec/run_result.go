package runspec

import (
	"errors"
	"fmt"
	. "github.com/elliotchance/orderedmap/v2"
	"github.com/fatih/color"
	"github.com/oleiade/lane/v2"
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	ozoneConfig "github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/logger_lib"
	"log"
	"strings"
)

type RunResult struct {
	RunId     string
	Status    CallstackStatus
	Roots     []*CallstackResultNode
	Index     *OrderedMap[string, *CallstackResultNode]
	listeners []UpdateListenerFunc
}

type CallstackStatus int

const Indent = "  "

const (
	NotStarted = iota
	Running
	Succeeded
	Failed
	Cached
)

type CallstackResultNode struct {
	Id          string
	logger      *logger_lib.Logger
	LogFile     string
	Caching     bool
	Hash        string
	Depth       int
	Root        *CallstackResultNode
	Children    []*CallstackResultNode
	Status      CallstackStatus
	Name        string
	Err         error
	IsCallstack bool
}

type UpdateListenerFunc func(*RunResult, bool)
type UpdateAllListenersFunc func(bool)

func NewRunResult() *RunResult {
	runResult := &RunResult{
		Status: Running,
		Index:  NewOrderedMap[string, *CallstackResultNode](),
	}

	return runResult
}

func (r *RunResult) ResetRunResult() {
	r.Status = Running
	r.Roots = nil
	r.Index = NewOrderedMap[string, *CallstackResultNode]()

	r.UpdateListeners(true)
}

func (r *RunResult) AddListener(listener UpdateListenerFunc) {
	r.listeners = append(r.listeners, listener)

	r.UpdateListeners(false)
}

//func (r *RunResult) HaveDescendantsSucceeded() {
//	for _, child := range r.Children {
//
//	}
//}

func (r *RunResult) findCallstackResultById(id string) (*CallstackResultNode, error) {
	// TODO this should probably use id instead of name
	for _, root := range r.Roots {
		stack := lane.NewStack[*CallstackResultNode]()
		stack.Push(root)

		for stack.Size() != 0 {
			current, _ := stack.Pop()

			if current.Id == id {
				return current, nil
			}

			// Push the children of the current node onto the stack in reverse order
			for i := len(current.Children) - 1; i >= 0; i-- {
				stack.Push(current.Children[i])
			}
		}
	}

	return nil, errors.New(fmt.Sprintf("CallstackResultNode id %s not found \n", id))
}

func (r *RunResult) GetLoggerForRunnableId(id string) (*logger_lib.Logger, error) {
	callstackResult, err := r.findCallstackResultById(id)
	if err != nil {
		return nil, err
	}

	return callstackResult.logger, nil
}

func (r *RunResult) SetRunnableHash(id, hash string) error {
	callstackResult, err := r.findCallstackResultById(id)
	if err != nil {
		return err
	}
	callstackResult.Hash = hash

	return nil
}

func getErrorMessage(node *CallstackResultNode) string {
	if node.Err != nil {
		return fmt.Sprintf("Error: %s", node.Err.Error())
	}

	return ""
}

func (r *RunResult) RunSpecRootNodeToRunResult(runId string, rootNode *RunspecRunnable, ozoneWorkDir string, config *ozoneConfig.OzoneConfig) {
	stack := lane.NewStack[*CallstackResultNode]()
	visited := make(map[*RunspecRunnable]*CallstackResultNode)

	rootLogger, err := logger_lib.New(runId, ozoneWorkDir, rootNode.GetRunnable().Name, config.Headless)
	if err != nil {
		log.Fatalln(err)
	}

	root := &CallstackResultNode{
		Id:          rootNode.GetRunnable().GetId(),
		logger:      rootLogger,
		LogFile:     rootLogger.GetLogFilePath(),
		Caching:     rootNode.GetRunnable().Cache,
		Depth:       0,
		Status:      NotStarted,
		Name:        rootNode.GetRunnable().Name,
		IsCallstack: true,
	}

	stack.Push(root)
	visited[rootNode] = root

	var asList []*CallstackResultNode

	asList = append(asList, root)

	for stack.Size() > 0 {
		// Pop the top node from the stack
		current, _ := stack.Pop()

		r.Index.Set(current.Id, current)

		// Get the corresponding original Node
		originalNode := getNodeById(rootNode, current.Id)

		callstackLogger, err := logger_lib.New(runId, ozoneWorkDir, current.Name, config.Headless)
		if err != nil {
			log.Fatalln(err)
		}

		// Copy children from the original Node to the current CallstackResultNode
		children := originalNode.GetChildren()
		current.Children = make([]*CallstackResultNode, len(children))
		for i, child := range children {
			if visited[child] == nil {
				isCallstack := false
				if child.HasCaching() {
					isCallstack = true
					callstackLogger, err = logger_lib.New(runId, ozoneWorkDir, child.GetRunnable().Name, config.Headless)
					if err != nil {
						log.Fatalln(err)
					}
				}
				childResult := &CallstackResultNode{
					Id:          child.GetRunnable().GetId(),
					logger:      callstackLogger,
					LogFile:     callstackLogger.GetLogFilePath(),
					Root:        root,
					Caching:     child.GetRunnable().Cache,
					Depth:       current.Depth + 1,
					Status:      NotStarted, // You can set the initial status as needed
					Name:        child.GetRunnable().Name,
					Err:         nil, // You can set the initial error as needed
					IsCallstack: isCallstack,
				}
				current.Children[i] = childResult
				stack.Push(childResult)
				visited[child] = childResult
			} else {
				current.Children[i] = visited[child]
			}
		}
	}

	r.Roots = append(r.Roots, root)
}

func getNodeById(current *RunspecRunnable, id string) *RunspecRunnable {
	if current.GetRunnable().GetId() == id {
		return current
	}

	for _, child := range current.GetChildren() {
		if node := getNodeById(child, id); node != nil {
			return node
		}
	}

	return nil
}

func (r *RunResult) AddRootCallstack(callstack *CallStack, logger *logger_lib.Logger) {
	r.Roots = append(r.Roots, &CallstackResultNode{
		Status:      NotStarted,
		Name:        callstack.RootRunnableName,
		logger:      logger,
		LogFile:     logger.GetLogFilePath(),
		IsCallstack: true,
	})
	r.UpdateListeners(false)
}

func (r *RunResult) UpdateListeners(reset bool) {
	for _, listener := range r.listeners {
		listener(r, reset)
	}
}

func (s CallstackStatus) String() string {
	return [...]string{"Not started", "Running", "Succeeded", "Failed", "Cached"}[s]
}

func (r *RunResult) AddCallstackResult(id string, status CallstackStatus, callStackErr error) error {
	if callStackErr != nil {
		r.Status = Failed
	}
	callstackResult, err := r.findCallstackResultById(id)
	if err != nil {
		return err
	} else {
		callstackResult.Status = status

		if status == Failed {
			callstackResult.Err = callStackErr
			if callstackResult.Root != nil {
				callstackResult.Root.Status = Failed
			}
		}
	}

	r.UpdateListeners(false)
	return nil
	// TODO if at this point, the runnable is caching, then we should update the cache if all children have succeeded.
}

func (r *RunResult) UpdateDaemonCacheResult(ozoneWorkDir string) {
	for _, root := range r.Roots {
		stack := lane.NewStack[*CallstackResultNode]()
		stack.Push(root)

		for stack.Size() != 0 {
			current, _ := stack.Pop()

			if current.Caching == true && current.Status == Succeeded {
				process_manager_client.CacheUpdate(ozoneWorkDir, current.Name, current.Hash)
				continue
			}

			for i := len(current.Children) - 1; i >= 0; i-- {
				stack.Push(current.Children[i])
			}
		}
	}
}

func (r *RunResult) PrintRunResult(print bool) string {
	s := fmt.Sprintln()

	for _, root := range r.Roots {

		stack := lane.NewStack[*CallstackResultNode]()

		if root != nil {
			stack.Push(root)
		}

		for stack.Size() != 0 {
			current, _ := stack.Pop()

			arrowChar := ""
			if current.Depth > 0 {
				arrowChar = "└-"
			}
			indent := fmt.Sprintf("%s%s", strings.Repeat(Indent, current.Depth), arrowChar)

			errorMessage := getErrorMessage(current)

			var statusColor func(a ...interface{}) string
			switch current.Status {
			case NotStarted:
				statusColor = color.New(color.FgBlue).SprintFunc()
			case Succeeded:
				statusColor = color.New(color.FgGreen).SprintFunc()
			case Failed:
				statusColor = color.New(color.FgRed).SprintFunc()
			default:
				statusColor = color.New(color.FgWhite).SprintFunc()
			}
			s += fmt.Sprintf("%s%s, Status: %s %s \n", indent, current.Name, statusColor(current.Status), errorMessage)

			// Push the children of the current node onto the stack in reverse order
			for i := len(current.Children) - 1; i >= 0; i-- {
				stack.Push(current.Children[i])
			}
		}

		if print == true {
			fmt.Print(s)
		}
	}

	return s
}

func (r *RunResult) AddSucceededCallstackResult(id string, err error) error {
	return r.AddCallstackResult(id, Succeeded, err)
}

func (r *RunResult) PrintIds() {
	log.Println("---===---")
	for _, root := range r.Roots {
		stack := lane.NewStack[*CallstackResultNode]()
		stack.Push(root)

		for stack.Size() != 0 {
			current, _ := stack.Pop()

			log.Printf("Id: %s, Name: %s, Status: %s \n", current.Id, current.Name, current.Status)

			// Push the children of the current node onto the stack in reverse order
			for i := len(current.Children) - 1; i >= 0; i-- {
				stack.Push(current.Children[i])
			}
		}
	}
}

func NewFailedCallstackResult(name string, err error, logger *logger_lib.Logger) *CallstackResultNode {
	return &CallstackResultNode{
		Status: Failed,
		Name:   name,
		logger: logger,
		Err:    err,
	}
}

func NewCachedCallstackResult(name string, logger *logger_lib.Logger) *CallstackResultNode {
	return &CallstackResultNode{
		Status: Cached,
		Name:   name,
		logger: logger,
	}
}

func (r *RunResult) PrintErrorLog() {
	for _, root := range r.Roots {
		stack := lane.NewStack[*CallstackResultNode]()
		stack.Push(root)

		if r.Status == Failed {
			fmt.Println("=================================================")
			fmt.Println("====================  Errors  ===================")
			fmt.Println("=================================================")
		}

		for stack.Size() != 0 {
			current, _ := stack.Pop()

			if current.Status == Failed {
				log.Printf("----------------------------------------------------------------------------\n")
				log.Printf("-                      Error logs for: %s                         \n", current.Name)
				log.Printf("----------------------------------------------------------------------------\n")
				lines, err := current.logger.TailFile(20)
				if err != nil {
					log.Fatalln(fmt.Sprintf("Error printing logFile for %s %s", current.Name, err))
				}
				for _, line := range lines {
					fmt.Printf("%s\n", line)
				}
			}

			// Push the children of the current node onto the stack in reverse order
			for i := len(current.Children) - 1; i >= 0; i-- {
				stack.Push(current.Children[i])
			}
		}
	}
}
