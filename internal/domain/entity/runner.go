package entity

// Runner represents a GitHub Actions self-hosted runner
type Runner struct {
	ID     int64
	Name   string
	Labels []string
	OS     string
	Status string
}
