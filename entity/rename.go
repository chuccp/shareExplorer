package entity

type Rename struct {
	Path     string `json:"path"`
	RootPath string `json:"rootPath"`
	OldName  string `json:"oldName"`
	NewName  string `json:"newName"`
}
