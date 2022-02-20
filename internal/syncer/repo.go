package syncer

// Repo defines the attributes of a github repository that will be
// required for the Syncer service.
type Repo struct {
	Owner     string
	Name      string
	CloneURL  string
	CommitSHA string
	LocalPath string
}
