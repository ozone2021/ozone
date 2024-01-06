package runspec

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/oleiade/lane/v2"
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	ozoneConfig "github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/logger_lib"
	"log"
	"strings"
)

type RunResult struct {
	Status    CallstackStatus
	Roots     []*CallstackResultNode
	index     map[string]*CallstackResultNode
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

type UpdateListenerFunc func(*RunResult)

func NewRunResult() *RunResult {
	runResult := &RunResult{
		Status: Running,
		index:  make(map[string]*CallstackResultNode),
	}

	return runResult
}

func (r *RunResult) AddListener(listener UpdateListenerFunc) {
	r.listeners = append(r.listeners, listener)

	r.UpdateListeners()
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

	return callstackResult.Logger, nil
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

func (r *RunResult) RunSpecRootNodeToRunResult(rootNode *RunspecRunnable, ozoneWorkDir string, config *ozoneConfig.OzoneConfig) {
	stack := lane.NewStack[*CallstackResultNode]()
	visited := make(map[*RunspecRunnable]*CallstackResultNode)

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

	r.Roots = append(r.Roots, root)
}

func getNodeByRunnableName(current *RunspecRunnable, name string) *RunspecRunnable {
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
	r.Roots = append(r.Roots, &CallstackResultNode{
		Status: Running,
		Name:   callstack.RootRunnableName,
		Logger: logger,
	})
	r.UpdateListeners()
}

func (r *RunResult) UpdateListeners() {
	for _, listener := range r.listeners {
		listener(r)
	}
}

func (s CallstackStatus) String() string {
	return [...]string{"Not started", "Running", "Succeeded", "Failed", "Cached"}[s]
}

func (r *RunResult) AddCallstackResult(id string, status CallstackStatus, err error) {
	if err != nil {
		r.Status = Failed
	}
	callstackResult, err := r.findCallstackResultById(id)

	if err == nil {
		callstackResult.Status = status
	} else {
		callstackResult.Err = err
		callstackResult.Status = Failed
	}

	r.UpdateListeners()
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

func (r *RunResult) AddSucceededCallstackResult(id string, err error) {
	r.AddCallstackResult(id, Succeeded, err)
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
}
