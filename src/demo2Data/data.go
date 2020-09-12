package demo2Data

type FileSystemAction struct {
	Path   string `json:"path"`
	File   string `json:"file"`
	Action string `json:"action"`
}

type Target struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type RawQuery struct {
	Database string `json:"database"`
	Query    string `json:"query"`
}

type MySqlUser struct {
	UUID      string `json:"uuid"`
	UserName  string `json:"userName"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Alias     string `json:"alias"`
	Password  string `json:"password"`
	Created   int64  `json:"created"`
	Modified  int64  `json:"modified"`
}

type Todo struct {
	UUID string `json:"uuid"`
}
