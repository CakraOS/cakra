package packagefmt

type Package struct {
	Name         string
	Version      string
	Release      int
	Architecture string
	Source       string
	Build        []string
	Install      []string
}
