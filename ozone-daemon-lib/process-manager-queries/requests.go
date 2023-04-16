package process_manager_queries

import . "github.com/ozone2021/ozone/ozone-lib/config/config_variable"

type ProcessCreateQuery struct {
	Name              string
	ProcessWorkingDir string
	OzoneWorkingDir   string
	Cmd               string
	Synchronous       bool
	IgnoreError       bool
	Env               *VariableMap
}

type DebugQuery struct {
	OzoneWorkingDir string
}

type ContextSetQuery struct {
	OzoneWorkingDir string
	Context         string
}

type IgnoreQuery struct {
	OzoneWorkingDir string
	Service         string
}

type HaltQuery struct {
	OzoneWorkingDir string
	Service         string
}

type DirQuery struct {
	OzoneWorkingDir string
}

type CacheQuery struct {
	OzoneWorkingDir     string
	RunnableName        string
	OzoneFileAndDirHash string
}

type StringReply struct {
	Body string
}

type BoolReply struct {
	Body bool
}
