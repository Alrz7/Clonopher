package composer

type FileInfo struct {
	Description string `yaml:"description"`
}

func (i *FileInfo) Render() string {
	if i == nil {
		return ""
	}
	return i.Description
}
