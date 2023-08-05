package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func releaseListHandler(collections *Collections) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
		c.Header("Access-Control-Max-Age", "15")
		cur, err := collections.Releases.Find(context.Background(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, ResponseStatus{"Unable to find releases"})
		}
		var result struct {
			Releases []LanguageRelease `json:"releases"`
		}
		for cur.Next(context.Background()) {
			var item bson.D
			if err := cur.Decode(&item); err != nil {
				log.Println("Error while decoding bson: ", err)
				continue
			}
			itemBytes, err := bson.Marshal(item)
			if err != nil {
				log.Println("Error converting item to json: ", err)
				continue
			}
			var rlsItem LanguageRelease
			bson.Unmarshal(itemBytes, &rlsItem)
			result.Releases = append(result.Releases, rlsItem)
		}
		c.JSON(http.StatusOK, result)
	}
}

func compileHandler(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
	c.Header("Access-Control-Max-Age", "15")
	var qatFile NewCompileFile
	err := c.BindJSON(qatFile)
	if err != nil {
		message := "Error decoding request body"
		log.Println(message)
		c.JSON(http.StatusBadRequest, ResponseStatus{message})
		return
	}
	if qatFile.ConfirmationKey != os.Getenv("CONFIRMATION_KEY") {
		message := "Source not confirmed"
		log.Println(message)
		c.JSON(http.StatusUnauthorized, ResponseStatus{message})
		return
	}
	uniq, err := uuid.NewUUID()
	if err != nil {
		message := "Cannot get UUID directory"
		log.Println(message)
		c.JSON(http.StatusInternalServerError, ResponseStatus{message})
		return
	}
	var dir string
	if len(os.Args) != 2 {
		dir = path.Join(os.Getenv("COMPILE_DIR"), uniq.String())
	} else {
		dir = path.Join(os.Args[1], os.Getenv("COMPILE_DIR"), uniq.String())
	}
	buildDir := path.Join(dir, "build")
	mainFile := path.Join(dir, "main.qat")
	err = os.MkdirAll(buildDir, 0755)
	if err != nil {
		message := "Cannot create build directory"
		log.Println(message)
		c.JSON(http.StatusInternalServerError, ResponseStatus{message})
		os.RemoveAll(dir)
		return
	}
	err = os.WriteFile(mainFile, []byte(qatFile.Content), 0755)
	if err != nil {
		message := "Cannot write contents to file for compile"
		log.Println(message)
		c.JSON(http.StatusInternalServerError, ResponseStatus{message})
		os.RemoveAll(dir)
		return
	}
	var cmd *exec.Cmd
	if len(os.Args) >= 3 {
		cmd = exec.Command(path.Join(os.Args[2], "qat"), "build", mainFile, "-o", buildDir, "--no-colors")
	} else {
		cmd = exec.Command("qat", "build", mainFile, "-o", buildDir, "--no-colors")
	}
	err = cmd.Run()
	if err != nil {
		message := "Running compiler failed: " + err.Error()
		log.Println(message)
		c.JSON(http.StatusInternalServerError, ResponseStatus{message})
		os.RemoveAll(dir)
		return
	}
	_, err = os.Stat(path.Join(dir, "build", "QatCompilationResult.json"))
	if err != nil {
		message := "Result file does not exist"
		log.Println(message)
		c.JSON(http.StatusInternalServerError, ResponseStatus{message})
		os.RemoveAll(dir)
		return
	}
	resContent, err := os.ReadFile(path.Join(dir, "build", "QatCompilationResult.json"))
	if err != nil {
		message := "Reading result file failed"
		log.Println(message)
		c.JSON(http.StatusInternalServerError, ResponseStatus{message})
		os.RemoveAll(dir)
		return
	}
	var sysCompRes SystemCompileResult
	err = json.Unmarshal(resContent, &sysCompRes)
	if err != nil {
		message := "Parsing result file failed"
		log.Println(message)
		c.JSON(http.StatusInternalServerError, ResponseStatus{message})
		os.RemoveAll(dir)
		return
	}
	if err == nil {
		log.Println("Writing final result")
		c.JSON(http.StatusOK, resContent)
		os.RemoveAll(dir)
		return
	} else {
		message := "Converting result failed"
		log.Println(message)
		c.JSON(http.StatusInternalServerError, ResponseStatus{message})
		return
	}
}

func downloadedReleaseHandler(collections *Collections) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
		c.Header("Access-Control-Max-Age", "15")
		var releaseDetails DownloadedReleaseDetails
		err := c.BindJSON(releaseDetails)
		if err != nil {
			message := "Error decoding request body"
			log.Println(message)
			c.JSON(http.StatusBadRequest, ResponseStatus{message})
			return
		}
		if releaseDetails.ConfirmationKey != os.Getenv("CONFIRMATION_KEY") {
			message := "Source not confirmed"
			log.Println(message)
			c.JSON(http.StatusNotAcceptable, ResponseStatus{message})
			return
		}
		var release LanguageRelease
		rlsResult := collections.Releases.FindOne(context.Background(), bson.M{"releaseID": releaseDetails.ReleaseID})
		err = rlsResult.Decode(&release)
		if err == nil {
			var foundPlatform bool
			var platformIndex int = 0
			for i := 0; i < len(release.Files); i++ {
				if release.Files[i].Id == releaseDetails.PlatformID {
					foundPlatform = true
					platformIndex = i
				}
			}
			if foundPlatform {
				updateRes, err := collections.Releases.UpdateOne(context.Background(),
					bson.M{"releaseID": releaseDetails.ReleaseID},
					bson.M{"$inc": bson.M{"files." + fmt.Sprint(platformIndex) + ".downloads": 1}})
				if err != nil || updateRes.ModifiedCount != 1 {
					message := "Could not update release"
					log.Println(message)
					c.JSON(http.StatusInternalServerError, ResponseStatus{message})
					return
				} else {
					message := "Updated release file download count successfully"
					log.Println(message)
					c.JSON(http.StatusOK, ResponseStatus{message})
					return
				}
			} else {
				message := "Platform not found"
				log.Println(message)
				c.JSON(http.StatusNotFound, ResponseStatus{message})
				return
			}
		} else {
			message := "No release found with ID"
			log.Println(message)
			c.JSON(http.StatusNotFound, ResponseStatus{message})
			return
		}
	}
}

func latestCommitHandler(collections *Collections) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
		c.Header("Access-Control-Max-Age", "15")
		var LimitVal int64 = 1
		cursor, err := collections.Commits.Find(context.Background(), bson.M{}, &options.FindOptions{Limit: &LimitVal, Sort: bson.M{"$natural": -1}})
		if err == nil {
			var item NewCommit
			for cursor.Next(context.Background()) {
				var itemBson bson.D
				if err := cursor.Decode(&itemBson); err != nil {
					message := "Error retrieving the latest commit from the fetched data"
					log.Println(message)
					c.JSON(http.StatusInternalServerError, ResponseStatus{message})
					return
				} else {
					itemBytes, err := bson.Marshal(itemBson)
					if err != nil {
						message := "Error converting latest commit BSON data to bytes"
						log.Println(message)
						c.JSON(http.StatusInternalServerError, ResponseStatus{message})
						return
					}
					err = bson.Unmarshal(itemBytes, &item)
					if err != nil {
						message := "Error converting bytes to latest commit data"
						log.Println(message)
						c.JSON(http.StatusInternalServerError, ResponseStatus{message})
						return
					}
					break
				}
			}
			c.JSON(http.StatusOK, item)
			return
		} else {
			message := "Error while looking for the latest commit"
			log.Println(message)
			c.JSON(http.StatusInternalServerError, ResponseStatus{message})
			return
		}
	}
}

func newCommitsHandler(collections *Collections) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
		c.Header("Access-Control-Max-Age", "15")
		var newCommitDetails PushedCommits
		err := c.BindJSON(newCommitDetails)
		if err != nil {
			message := "Could not convert request body to JSON"
			log.Println(message)
			c.JSON(http.StatusInternalServerError, ResponseStatus{message})
			return
		}
		if newCommitDetails.ConfirmationKey != os.Getenv("CONFIRMATION_KEY") {
			message := "Source not confirmed"
			log.Println(message)
			c.JSON(http.StatusNotAcceptable, ResponseStatus{message})
			return
		}
		var commitVals bson.A
		for i := 0; i < len(newCommitDetails.Commits); i++ {
			commitVals = append(commitVals, bson.M{
				"id":         newCommitDetails.Commits[i].Id,
				"title":      newCommitDetails.Commits[i].Title,
				"message":    newCommitDetails.Commits[i].Message,
				"author":     newCommitDetails.Commits[i].Author,
				"repository": newCommitDetails.Commits[i].Repository,
				"site":       newCommitDetails.Commits[i].Site,
				"timestamp":  newCommitDetails.Commits[i].Timestamp,
				"ref":        newCommitDetails.Commits[i].Ref,
			})
		}
		_, err = collections.Commits.InsertMany(context.Background(), commitVals)
		if err != nil {
			message := "Could not add commits to the database"
			log.Println(message)
			c.JSON(http.StatusInternalServerError, ResponseStatus{message})
			return
		}
		c.JSON(http.StatusOK, ResponseStatus{"Added commits successfully"})
	}
}

func releaseCountHandler(collections *Collections) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
		c.Header("Access-Control-Max-Age", "15")
		count, err := collections.Releases.CountDocuments(context.Background(), bson.M{})
		if err == nil {
			c.JSON(http.StatusOK, CommitCount{Count: count})
		} else {
			message := "Could not retrieve number of releases"
			log.Println(message)
			c.JSON(http.StatusInternalServerError, ResponseStatus{message})
		}
	}
}

func projectStatsHandler(collections *Collections) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
		c.Header("Access-Control-Max-Age", "15")
		serverConfigRes := collections.Commits.FindOne(context.Background(), bson.M{})
		var config ServerConfig
		err := serverConfigRes.Decode(config)
		if err == nil {
			wakatimeBaseUrl := "https://wakatime.com/api/v1/users/current/all_time_since_today?project="
			var allStats AllStatsResult
			var reqBytes []byte
			client := http.Client{Timeout: 30 * time.Second}
			projectHandler := func(projectName string) (*ProjectStats, error) {
				projectRequest, err := http.NewRequestWithContext(context.Background(), http.MethodGet, wakatimeBaseUrl+projectName, bytes.NewReader(reqBytes))
				if err != nil {
					message := "could not create request for stats of the compiler project"
					log.Println(message)
					return nil, errors.New(message)
				}
				projectRequest.Header.Set("Authorization", "Bearer "+config.Wakatime.AccessToken)
				projectRequest.Header.Set("Access-Control-Origin-Policy", "*")
				resp, err := client.Do(projectRequest)
				if err != nil {
					message := "error making request for stats of the compiler project"
					log.Println(message)
					return nil, errors.New(message)
				}
				var resBytes []byte
				_, err = resp.Body.Read(resBytes)
				if err != nil {
					message := "error reading response for stats of the compiler project"
					log.Println(message)
					return nil, errors.New(message)
				}
				result := new(ProjectStats)
				err = json.Unmarshal(resBytes, result)
				if err != nil {
					message := "error decoding stats of the compiler project to JSON"
					log.Println(message)
					return nil, errors.New(message)
				}
				return result, nil
			}
			compilerProjectStats, err := projectHandler("qat")
			if err != nil {
				c.JSON(http.StatusInternalServerError, ResponseStatus{err.Error()})
				return
			}
			siteProjectStats, err := projectHandler("qatdev")
			if err != nil {
				c.JSON(http.StatusInternalServerError, ResponseStatus{err.Error()})
				return
			}
			serverProjectStats, err := projectHandler("QatDevServer")
			if err != nil {
				c.JSON(http.StatusInternalServerError, ResponseStatus{err.Error()})
				return
			}
			vscodeExtProjectStats, err := projectHandler("qat_vscode")
			if err != nil {
				c.JSON(http.StatusInternalServerError, ResponseStatus{err.Error()})
				return
			}
			docsProjectStats, err := projectHandler("QatDocs")
			if err != nil {
				c.JSON(http.StatusInternalServerError, ResponseStatus{err.Error()})
				return
			}
			allStats.Compiler = *compilerProjectStats
			allStats.Website = *siteProjectStats
			allStats.Server = *serverProjectStats
			allStats.VSCode = *vscodeExtProjectStats
			allStats.Docs = *docsProjectStats
			c.JSON(http.StatusOK, allStats)
		} else {
			message := "Could not decode server config"
			log.Println(message)
			c.JSON(http.StatusInternalServerError, ResponseStatus{message})
		}
	}
}
