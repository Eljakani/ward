package models

// ProjectContext holds resolved project metadata that scanners consume.
type ProjectContext struct {
	RootPath          string
	LaravelVersion    string
	PHPVersion        string
	ProjectName       string
	FrameworkType     string
	ComposerDeps      map[string]string // from composer.json require
	InstalledPackages map[string]string // from composer.lock (resolved versions)
	EnvVariables      map[string]string
	ConfigFiles       []string
}
