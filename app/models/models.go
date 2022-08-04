package models

// ClientData - struct for user data implementation
type ClientData struct {
	Cookie string      `json:"cookie"` // Cookie - user cookie
	Key    string      `json:"key"`    // Key - cookie sign key
	Short  []ShortData `json:"values"` //  Short - short urls
}

// ShortData - struct for short url user storage implementation
type ShortData struct {
	Short   string `json:"short"`   // Short - short url
	Long    string `json:"long"`    // Long - original url
	Deleted bool   `json:"deleted"` // Deleted - current short url status
}

// DelWorker - struct for delete worker input
type DelWorker struct {
	Cookie string   // Cookie - user identification
	Tags   []string // Tags - list of url tags
}

// DelTask - struct atomic for delete worker
type DelTask struct {
	Cookie string // Cookie - user identification
	Tag    string // Short url tag
}
