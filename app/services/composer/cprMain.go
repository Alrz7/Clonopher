package composer

type FileInfo struct {
	Description string `yaml:"description"`
}

type StageInfo struct {
	Description string `yaml:"description"`
	// Future:
	// Tags
	// Summary
	// History
	// ...
}

func (i *FileInfo) Join() string {
	if i == nil {
		return ""
	}
	return i.Description
}

func (i *StageInfo) Join() string {
	if i == nil {
		return ""
	}
	return i.Description
}
