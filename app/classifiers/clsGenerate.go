package classifiers

type ObjectInfo struct {
	Description string `yaml:"description"`
}

func (i *ObjectInfo) Render() string {
	if i == nil {
		return ""
	}
	return i.Description
}
