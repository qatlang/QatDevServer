package main

import "go.mongodb.org/mongo-driver/mongo"

type Collections struct {
	Updates  *mongo.Collection
	Releases *mongo.Collection
	Commits  *mongo.Collection
	Config   *mongo.Collection
}

type WakatimeConfig struct {
	AccessToken  string `json:"accessToken" bson:"accessToken"`
	RefreshToken string `json:"refreshToken" bson:"refreshToken"`
	ExpiresAt    string `json:"expiresAt" bson:"expiresAt"`
	ClientSecret string `json:"clientSecret" bson:"clientSecret"`
	ClientID     string `json:"clientID" bson:"clientID"`
	RefreshURL   string `json:"refreshURL" bson:"refreshURL"`
}

type ServerConfig struct {
	Wakatime WakatimeConfig `json:"wakatime" bson:"wakatime"`
}

type LanguageRelease struct {
	ReleaseID string `json:"releaseID"`
	Version   struct {
		Value        string `json:"value"`
		IsPrerelease bool   `json:"isPrerelease"`
		Prerelease   string `json:"prerelease"`
	} `json:"version"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Files   []struct {
		Id        string `json:"id"`
		Platform  string `json:"platform"`
		Target    string `json:"target"`
		Downloads int    `json:"downloads"`
		Path      string `json:"path"`
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
	IsError  bool      `json:"isError"`
	Message  string    `json:"message"`
	HasRange bool      `json:"hasRange"`
	Range    FileRange `json:"range,omitempty"`
}

type DownloadedReleaseDetails struct {
	ConfirmationKey string `json:"confirmationKey"`
	ReleaseID       string `json:"releaseID"`
	PlatformID      string `json:"platformID"`
}

type PushedCommits struct {
	ConfirmationKey string      `json:"confirmationKey"`
	Commits         []NewCommit `json:"commits"`
}

type NewCommit struct {
	Id      string `json:"id"`
	Title   string `json:"title"`
	Message string `json:"message"`
	Author  struct {
		Name  string `json:"name"`
		Email string `json:"email,omitempty"`
	}
	Repository string `json:"repository"`
	Site       string `json:"site"`
	Timestamp  string `json:"timestamp"`
	Ref        string `json:"ref"`
}

type CommitCount struct {
	Count int64 `json:"count"`
}

type SystemCompileResult struct {
	Problems        []Problem `json:"problems"`
	Status          bool      `json:"status"`
	CompilationTime int64     `json:"compilationTime"`
	LinkingTime     int64     `json:"linkingTime"`
	BinarySizes     []int64   `json:"binarySizes"`
	HasMain         bool      `json:"hasMain"`
}

type ResponseStatus struct {
	Status string `json:"status"`
}

type ProjectStats struct {
	Data struct {
		Decimal           string  `json:"decimal"`
		Digital           string  `json:"digital"`
		IsUpToDate        bool    `json:"is_up_to_date"`
		PercentCalculated float64 `json:"percent_calculated"`
		Project           string  `json:"project"`
		Range             struct {
			End       string `json:"end"`
			EndDate   string `json:"end_date"`
			EndText   string `json:"end_text"`
			Start     string `json:"start"`
			StartDate string `json:"start_date"`
			StartText string `json:"start_text"`
			Timezone  string `json:"timezone"`
		} `json:"range"`
		Text         string  `json:"text"`
		Timeout      int64   `json:"timeout"`
		TotalSeconds float64 `json:"total_seconds"`
	} `json:"data"`
}

type AllStatsResult struct {
	Compiler ProjectStats `json:"compiler"`
	Website  ProjectStats `json:"website"`
	Server   ProjectStats `json:"server"`
	VSCode   ProjectStats `json:"vscode"`
	Docs     ProjectStats `json:"docs"`
	Tom      ProjectStats `json:"tom"`
}
