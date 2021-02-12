package github

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type NotGitHubURLErr struct{}

func (NotGitHubURLErr) Error() string {
	return "not a GitHub URL"
}

func parseGitHubURL(v string) ([]string, error) {
	u, err := url.Parse(v)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse URL")
	}

	if u.Host != "github.com" {
		return nil, NotGitHubURLErr{}
	}

	return strings.Split(strings.TrimPrefix(u.Path, "/"), "/"), nil
}

// GitHub Repo URLs look like
// https://github.com/akitasoftware/superstar
func ParseRepoURL(u string) (*Repo, error) {
	parts, err := parseGitHubURL(u)
	if err != nil {
		return nil, err
	} else if len(parts) < 2 {
		return nil, errors.Errorf("not a valid GitHub repo URL: %s", u)
	}

	return &Repo{
		Owner: parts[0],
		Name:  parts[1],
	}, nil
}

// GitHub PR URLs look like
// https://github.com/akitasoftware/superstar/pull/678
func ParsePullRequestURL(u string) (*PullRequest, error) {
	parts, err := parseGitHubURL(u)
	if err != nil {
		return nil, err
	} else if len(parts) < 4 {
		return nil, errors.Errorf("not a valid GitHub PR URL: %s", u)
	} else if parts[2] != "pull" {
		return nil, errors.Errorf("not a valid GitHub PR URL: %s", u)
	}

	repoOwner, repoName := parts[0], parts[1]
	prNum, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, errors.Wrapf(err, `Failed to parse GitHub PR number, value="%s"`, parts[3])
	}

	return &PullRequest{
		Repo: Repo{
			Owner: repoOwner,
			Name:  repoName,
		},
		Num: prNum,
	}, nil
}
