package gitlab

// Info about merge request.
type MRInfo struct {
	Project string `json:"project"`

	// GitLab MR IID
	// https://docs.gitlab.com/ee/api/#id-vs-iid
	IID string `json:"iid"`

	Branch string `json:"branch"`
	Commit string `json:"commit"`
}
