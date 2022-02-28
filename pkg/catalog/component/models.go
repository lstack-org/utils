package component

import "time"

type Status string

// 组件状态
const (
	ComponentAvailable    Status = "Available"
	ComponentNotAvailable Status = "Unavailable"
	ComponentCreating     Status = "Creating"
	ComponentUpdating     Status = "Updating"
)

type Component struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	CreateTime        time.Time `json:"createTime"`
	AgentName         string    `json:"agentName"`
	ProviderName      string    `json:"providerName"`
	GroupId           int64     `json:"groupId"`
	GroupName         string    `json:"groupName"`
	GroupFullPath     string    `json:"groupFullPath"`
	RepositoryId      int64     `json:"repositoryId"`
	RepositoryName    string    `json:"repositoryName"`
	CodeRepositoryUrl string    `json:"codeRepositoryUrl"`
	State             Status    `json:"status"`
	StatusDescription string    `json:"statusDescription"`
}

type EditArg struct {
	OldComponent Component `json:"oldComponent"`
	NewComponent Component `json:"newComponent"`
}

type DeleteArg struct {
	Components []Component `json:"components"`
}
