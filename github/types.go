package github

type Repo struct {
	Owner string
	Name  string
}

type PullRequest struct {
	Repo   Repo
	Num    int
	Branch string
	Commit string
}

// Used in CreateSpec messages
type PRInfo struct {
	RepoOwner string `json:"repo_owner"`
	RepoName  string `json:"repo_name"`
	Num       int    `json:"num"`
}
