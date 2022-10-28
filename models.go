package main

import "go.mongodb.org/mongo-driver/mongo"

type Collections struct {
	Updates  *mongo.Collection
	Releases *mongo.Collection
}

type ReleaseVersionInfo struct {
	Value        string `json:"value"`
	IsPrerelease string `json:"isPrerelease"`
	Prerelease   string `json:"prerelease"`
}

type ReleaseArtefact struct {
	Platform     string `json:"platform"`
	Architecture string `json:"architecture"`
	Path         string `json:"path"`
}

type LanguageRelease struct {
	Version   ReleaseVersionInfo `json:"version"`
	Title     string             `json:"title"`
	Content   string             `json:"content"`
	Files     []ReleaseArtefact  `json:"files"`
	Index     int                `json:"index"`
	CreatedAt string             `json:"createdAt"`
}

type LanguageUpdate struct {
	Content   string `json:"content"`
	Title     string `json:"title"`
	CreatedAt string `json:"createdAt"`
	Index     int    `json:"index"`
}

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
