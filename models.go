package main

import "go.mongodb.org/mongo-driver/mongo"

type Collections struct {
	Updates  *mongo.Collection
	Releases *mongo.Collection
}

type LanguageRelease struct {
	Version struct {
		Value        string `json:"value"`
		IsPrerelease bool   `json:"isPrerelease"`
		Prerelease   string `json:"prerelease"`
	} `json:"version"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Files   []struct {
		Platform     string `json:"platform"`
		Architecture string `json:"architecture"`
		Downloads    int    `json:"downloads"`
		Path         string `json:"path"`
	} `json:"files"`
	Index     int    `json:"index"`
	CreatedAt string `json:"createdAt"`
}

type LanguageUpdate struct {
	Content   string `json:"content"`
	Title     string `json:"title"`
	CreatedAt string `json:"createdAt"`
	Index     int    `json:"index"`
}

type NewCompileFile struct {
	ConfirmationKey string `json:"confirmationKey"`
	Content         string `json:"content"`
	Time            string `json:"time"`
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

type StatusFail struct {
	Status string `json:"status"`
}
