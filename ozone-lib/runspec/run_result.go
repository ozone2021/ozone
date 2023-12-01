package runspec

import (
	"errors"
	"fmt"
	"github.com/oleiade/lane/v2"
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	ozoneConfig "github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/logger_lib"
	"log"
	"strings"
)

type RunResult struct {
	Status CallstackStatus
	Root   *CallstackResultNode
	index  map[string]*CallstackResultNode
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
	Id       string
	Logger   *logger_lib.Logger
	Caching  bool
	Hash     string
	Depth    int
	Children []*CallstackResultNode
	Status   CallstackStatus
	Name     string
	Err      error
}

func NewRunResult() *RunResult {
	return &RunResult{
		Status: Running,
		index:  make(map[string]*CallstackResultNode),
	}
}

func (r *RunResult) findCallstackResult(name string) (*CallstackResultNode, error) {
	stack := lane.NewStack[*CallstackResultNode]()
	stack.Push(r.Root)

	for stack.Size() != 0 {
		current, _ := stack.Pop()

		if current.Name == name {
			return current, nil
		}

		// Push the children of the current node onto the stack in reverse order
		for i := len(current.Children) - 1; i >= 0; i-- {
			stack.Push(current.Children[i])
		}
	}

	return nil, errors.New(fmt.Sprintf("CallstackResultNode %s not found \n", name))
}

func (r *RunResult) GetLoggerForRunnable(name string) (*logger_lib.Logger, error) {
	callstackResult, err := r.findCallstackResult(name)
	if err != nil {
		return nil, err
	}

	return callstackResult.Logger, nil
}

func (r *RunResult) SetRunnableHash(name, hash string) error {
	callstackResult, err := r.findCallstackResult(name)
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

func (r *RunResult) RunSpecRootNodeToRunResult(rootNode Node, ozoneWorkDir string, config *ozoneConfig.OzoneConfig) {
	stack := lane.NewStack[*CallstackResultNode]()
	visited := make(map[Node]*CallstackResultNode)

	rootLogger, err := logger_lib.New(ozoneWorkDir, rootNode.GetRunnable().Name, config.Headless)
	if err != nil {
		log.Fatalln(err)
	}

	root := &CallstackResultNode{
		Id:      rootNode.GetRunnable().GetId(),
		Logger:  rootLogger,
		Caching: rootNode.GetRunnable().Cache,
		Depth:   0,
		Status:  Running, // You can set the initial status as needed
		Name:    rootNode.GetRunnable().Name,
	}

	stack.Push(root)
	visited[rootNode] = root

	for stack.Size() > 0 {
		// Pop the top node from the stack
		current, _ := stack.Pop()

		// Get the corresponding original Node
		originalNode := getNodeByRunnableName(rootNode, current.Name)

		callstackLogger, err := logger_lib.New(ozoneWorkDir, current.Name, config.Headless)
		if err != nil {
			log.Fatalln(err)
		}

		// Copy children from the original Node to the current CallstackResultNode
		children := originalNode.GetChildren()
		current.Children = make([]*CallstackResultNode, len(children))
		for i, child := range children {
			if visited[child] == nil {
				if child.HasCaching() {
					callstackLogger, err = logger_lib.New(ozoneWorkDir, child.GetRunnable().Name, config.Headless)
					if err != nil {
						log.Fatalln(err)
					}
				}
				childResult := &CallstackResultNode{
					Id:      child.GetRunnable().GetId(),
					Logger:  callstackLogger,
					Caching: child.GetRunnable().Cache,
					Depth:   current.Depth + 1,
					Status:  NotStarted, // You can set the initial status as needed
					Name:    child.GetRunnable().Name,
					Err:     nil, // You can set the initial error as needed
				}
				current.Children[i] = childResult
				stack.Push(childResult)
				visited[child] = childResult
			} else {
				current.Children[i] = visited[child]
			}
		}
	}

	r.Root = root
}

func getNodeByRunnableName(current Node, name string) Node {
	if current.GetRunnable().Name == name {
		return current
	}

	for _, child := range current.GetChildren() {
		if node := getNodeByRunnableName(child, name); node != nil {
			return node
		}
	}

	return nil
}

func (r *RunResult) AddRootCallstack(callstack *CallStack, logger *logger_lib.Logger) {
	r.Root = &CallstackResultNode{
		Status: Running,
		Name:   callstack.RootRunnableName,
		Logger: logger,
	}
}

func (s CallstackStatus) String() string {
	return [...]string{"Not started", "Running", "Succeeded", "Failed", "Cached"}[s]
}

func (r *RunResult) AddCallstackResult(runnableName string, status CallstackStatus, err error) {
	if err != nil {
		r.Status = Failed
	}
	callstackResult, err := r.findCallstackResult(runnableName)

	if err == nil {
		callstackResult.Status = status
	} else {
		callstackResult.Err = err
		callstackResult.Status = Failed
	}
	// TODO if at this point, the runnable is caching, then we should update the cache if all children have succeeded.
}

func (r *RunResult) UpdateDaemonCacheResult(ozoneWorkDir string) {
	stack := lane.NewStack[*CallstackResultNode]()
	stack.Push(r.Root)

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

func (r *RunResult) PrintRunResult() {
	stack := lane.NewStack[*CallstackResultNode]()
	stack.Push(r.Root)

	for stack.Size() != 0 {
		current, _ := stack.Pop()

		indent := strings.Repeat(Indent, current.Depth)

		errorMessage := getErrorMessage(current)

		fmt.Printf("%sCallstack: %s, Status: %s %s \n", indent, current.Name, current.Status, errorMessage)

		// Push the children of the current node onto the stack in reverse order
		for i := len(current.Children) - 1; i >= 0; i-- {
			stack.Push(current.Children[i])
		}
	}
}

func NewSucceededCallstackResult(name string, logger *logger_lib.Logger) *CallstackResultNode {
	return &CallstackResultNode{
		Status: Succeeded,
		Name:   name,
		Logger: logger,
	}
}

func NewFailedCallstackResult(name string, err error, logger *logger_lib.Logger) *CallstackResultNode {
	return &CallstackResultNode{
		Status: Failed,
		Name:   name,
		Logger: logger,
		Err:    err,
	}
}

func NewCachedCallstackResult(name string, logger *logger_lib.Logger) *CallstackResultNode {
	return &CallstackResultNode{
		Status: Cached,
		Name:   name,
		Logger: logger,
	}
}

func (r *RunResult) PrintErrorLog() {
	stack := lane.NewStack[*CallstackResultNode]()
	stack.Push(r.Root)

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
			lines, err := current.Logger.TailFile(20)
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
