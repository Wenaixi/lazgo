package models

// FileInfo represents a file or folder item from listing.
type FileInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Size     string `json:"size"`
	Time     string `json:"time"`
	IsFolder bool   `json:"is_folder"`
}

// FolderInfo represents folder detail.
type FolderInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ShareLink represents a share link result.
type ShareLink struct {
	URL         string `json:"url"`
	HasPassword bool   `json:"has_password"`
	Password    string `json:"password"`
	ShareID     string `json:"share_id"`
}

// DirectLink represents a direct download link.
type DirectLink struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Method   string `json:"method"`
	Warning  string `json:"warning,omitempty"`
}

// LoginResult represents a login result.
type LoginResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
