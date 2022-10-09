package main

type NewCompileFile struct {
	Content string `json:"content"`
	Time    string `json:"time"`
}

type FilePos struct {
	Line int `json:"line"`
	Char int `json:"char"`
}

type FileRange struct {
	File  string  `json:"file"`
	Start FilePos `json:"start"`
	End   FilePos `json:"end"`
}

type Problem struct {
	IsError bool      `json:"isError"`
	Message string    `json:"message"`
	Range   FileRange `json:"range"`
}

type SystemCompileResult struct {
	Problems  []Problem `json:"problems"`
	Status    bool      `json:"status"`
	QatTime   int       `json:"qatTime"`
	ClangTime int       `json:"clangTime"`
	HasMain   bool      `json:"hasMain"`
}

type CompileStatusFail struct {
	Status string `json:"status"`
}
