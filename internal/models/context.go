package models

// ProjectContext holds resolved project metadata that scanners consume.
type ProjectContext struct {
	RootPath       string
	LaravelVersion string
	PHPVersion     string
	ProjectName    string
	FrameworkType  string
	ComposerDeps   map[string]string
	EnvVariables   map[string]string
	ConfigFiles    []string
}
