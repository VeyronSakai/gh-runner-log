package debug

import "strings"

// matchScope verifies that the repository string should be included for the given owner/repo/org filters.
func matchScope(repository, owner, repo, org string) bool {
	if org != "" {
		return strings.HasPrefix(repository, org+"/")
	}
	if owner == "" || repo == "" {
		return true
	}
	return repository == owner+"/"+repo
}
