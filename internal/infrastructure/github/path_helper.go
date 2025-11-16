package github

import "fmt"

// getActionsBasePath returns the base path for GitHub Actions API
// Returns "orgs/{org}/actions" for organization scope or "repos/{owner}/{repo}/actions" for repository scope
func getActionsBasePath(owner, repo, org string) string {
	if org != "" {
		return fmt.Sprintf("orgs/%s/actions", org)
	}
	return fmt.Sprintf("repos/%s/%s/actions", owner, repo)
}

// getRepoActionsBasePath returns the base path for repository-scoped GitHub Actions API
// Always returns "repos/{owner}/{repo}/actions" regardless of org
func getRepoActionsBasePath(owner, repo string) string {
	return fmt.Sprintf("repos/%s/%s/actions", owner, repo)
}
