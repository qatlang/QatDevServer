package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func releaseListHandler(collections *Collections) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			{
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
				w.Header().Set("Access-Control-Max-Age", "15")
				cur, err := collections.Releases.Find(context.TODO(), bson.M{})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					statusData, err := json.Marshal(ResponseStatus{"Unable to find releases"})
					if err == nil {
						w.Write(statusData)
					}
				}
				var result struct {
					Releases []LanguageRelease `json:"releases"`
				}
				for cur.Next(context.TODO()) {
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
				resultBytes, err := json.Marshal(result)
				if err != nil {
					log.Println("Error while converting result to json: ", err)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write(resultBytes)
				break
			}
		default:
			{
				w.WriteHeader(http.StatusNotFound)
				var status ResponseStatus
				status.Status = "Invalid request"
				statusBytes, err := json.Marshal(status)
				if err != nil {
					return
				}
				w.Write(statusBytes)
				break
			}
		}
	}
}

func writingStatusFailed(message string) {
	log.Println("WRITING STATUS FAILED :: " + message)
}

func compileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		{
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
			w.Header().Set("Access-Control-Max-Age", "15")
			var qatFile NewCompileFile
			json.NewDecoder(r.Body).Decode(&qatFile)
			if qatFile.ConfirmationKey != os.Getenv("CONFIRMATION_KEY") {
				w.WriteHeader(http.StatusNotAcceptable)
				message := "Source not confirmed"
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
				return
			}
			uniq, err := uuid.NewUUID()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				message := "Cannot get UUID directory"
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
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
				w.WriteHeader(http.StatusInternalServerError)
				message := "Cannot create build directory"
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
				os.RemoveAll(dir)
				return
			}
			err = os.WriteFile(mainFile, []byte(qatFile.Content), 0755)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				message := "Cannot write contents to file for compile"
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
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
				w.WriteHeader(http.StatusInternalServerError)
				message := "Running compiler failed: " + err.Error()
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
				os.RemoveAll(dir)
				return
			}
			_, err = os.Stat(path.Join(dir, "build", "QatCompilationResult.json"))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				message := "Result file does not exist"
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
				os.RemoveAll(dir)
				return
			}
			resContent, err := os.ReadFile(path.Join(dir, "build", "QatCompilationResult.json"))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				message := "Reading result file failed"
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
				os.RemoveAll(dir)
				return
			}
			var sysCompRes SystemCompileResult
			err = json.Unmarshal(resContent, &sysCompRes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				message := "Parsing result file failed"
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
				os.RemoveAll(dir)
				return
			}
			if err == nil {
				log.Println("Writing final result")
				w.Write(resContent)
				os.RemoveAll(dir)
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				message := "Converting result failed"
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
				return
			}
		}
	default:
		{
			w.WriteHeader(http.StatusNotFound)
			message := "Invalid method"
			status, err := json.Marshal(ResponseStatus{message})
			if err == nil {
				log.Println(message)
				w.Write(status)
			} else {
				writingStatusFailed(message)
			}
			break
		}
	}
}

func downloadedReleaseHandler(collections *Collections) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			{
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
				w.Header().Set("Access-Control-Max-Age", "15")
				var releaseDetails DownloadedReleaseDetails
				json.NewDecoder(r.Body).Decode(&releaseDetails)
				if releaseDetails.ConfirmationKey != os.Getenv("CONFIRMATION_KEY") {
					w.WriteHeader(http.StatusNotAcceptable)
					message := "Source not confirmed"
					status, err := json.Marshal(ResponseStatus{message})
					if err == nil {
						log.Println(message)
						w.Write(status)
					} else {
						writingStatusFailed(message)
					}
					return
				}
				var release LanguageRelease
				rlsResult := collections.Releases.FindOne(context.TODO(), bson.M{"releaseID": releaseDetails.ReleaseID})
				err := rlsResult.Decode(&release)
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
						updateRes, err := collections.Releases.UpdateOne(context.TODO(),
							bson.M{"releaseID": releaseDetails.ReleaseID},
							bson.M{"$inc": bson.M{"files." + fmt.Sprint(platformIndex) + ".downloads": 1}})
						if err != nil || updateRes.ModifiedCount != 1 {
							w.WriteHeader(http.StatusInternalServerError)
							message := "Could not update release"
							status, err := json.Marshal(ResponseStatus{message})
							if err == nil {
								log.Println(message)
								w.Write(status)
							} else {
								writingStatusFailed(message)
							}
							return
						} else {
							w.WriteHeader(http.StatusOK)
							message := "Updated release file download count successfully"
							status, err := json.Marshal(ResponseStatus{message})
							if err == nil {
								log.Println(message)
								w.Write(status)
							} else {
								writingStatusFailed(message)
							}
							return
						}
					} else {
						w.WriteHeader(http.StatusNotFound)
						message := "Platform not found"
						status, err := json.Marshal(ResponseStatus{message})
						if err == nil {
							log.Println(message)
							w.Write(status)
						} else {
							writingStatusFailed(message)
						}
						return
					}
				} else {
					w.WriteHeader(http.StatusNotFound)
					message := "No release found with ID"
					status, err := json.Marshal(ResponseStatus{message})
					if err == nil {
						log.Println(message)
						w.Write(status)
					} else {
						writingStatusFailed(message)
					}
					return
				}
			}
		default:
			{
				w.WriteHeader(http.StatusNotFound)
				message := "Invalid method"
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
				break
			}
		}
	}
}

func latestCommitHandler(collections *Collections) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			{
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
				w.Header().Set("Access-Control-Max-Age", "15")
				var LimitVal int64 = 1
				cursor, err := collections.Commits.Find(context.TODO(), bson.M{}, &options.FindOptions{Limit: &LimitVal, Sort: bson.M{"$natural": -1}})
				if err == nil {
					var item NewCommit
					for cursor.Next(context.TODO()) {
						var itemBson bson.D
						if err := cursor.Decode(&itemBson); err != nil {
							w.WriteHeader(http.StatusInternalServerError)
							message := "Error retrieving the latest commit from the fetched data"
							status, err := json.Marshal(ResponseStatus{message})
							if err == nil {
								log.Println(message)
								w.Write(status)
							} else {
								writingStatusFailed(message)
							}
							return
						} else {
							itemBytes, err := bson.Marshal(itemBson)
							if err != nil {
								w.WriteHeader(http.StatusInternalServerError)
								message := "Error converting latest commit BSON data to bytes"
								status, err := json.Marshal(ResponseStatus{message})
								if err == nil {
									log.Println(message)
									w.Write(status)
								} else {
									writingStatusFailed(message)
								}
								return
							}
							err = bson.Unmarshal(itemBytes, &item)
							if err != nil {
								w.WriteHeader(http.StatusInternalServerError)
								message := "Error converting bytes to latest commit data"
								status, err := json.Marshal(ResponseStatus{message})
								if err == nil {
									log.Println(message)
									w.Write(status)
								} else {
									writingStatusFailed(message)
								}
								return
							}
							break
						}
					}
					itemBytes, err := json.Marshal(item)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						message := "Error converting latest commit data to JSON bytes"
						status, err := json.Marshal(ResponseStatus{message})
						if err == nil {
							log.Println(message)
							w.Write(status)
						} else {
							writingStatusFailed(message)
						}
						return
					}
					w.WriteHeader(http.StatusOK)
					w.Write(itemBytes)
					return
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					message := "Error while looking for the latest commit"
					status, err := json.Marshal(ResponseStatus{message})
					if err == nil {
						log.Println(message)
						w.Write(status)
					} else {
						writingStatusFailed(message)
					}
					return
				}
			}
		default:
			{
				w.WriteHeader(http.StatusNotFound)
				message := "Invalid method"
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
				break
			}
		}
	}
}

func newCommitsHandler(collections *Collections) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			{
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
				w.Header().Set("Access-Control-Max-Age", "15")
				var newCommitDetails PushedCommits
				err := json.NewDecoder(r.Body).Decode(&newCommitDetails)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					message := "Could not convert request body to JSON"
					status, err := json.Marshal(ResponseStatus{message})
					if err == nil {
						log.Println(message)
						w.Write(status)
					} else {
						writingStatusFailed(message)
					}
					return
				}
				if newCommitDetails.ConfirmationKey != os.Getenv("CONFIRMATION_KEY") {
					w.WriteHeader(http.StatusNotAcceptable)
					message := "Source not confirmed"
					status, err := json.Marshal(ResponseStatus{message})
					if err == nil {
						log.Println(message)
						w.Write(status)
					} else {
						writingStatusFailed(message)
					}
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
				_, err = collections.Commits.InsertMany(context.TODO(), commitVals)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					message := "Could not add commits to the database"
					status, err := json.Marshal(ResponseStatus{message})
					if err == nil {
						log.Println(message)
						w.Write(status)
					} else {
						writingStatusFailed(message)
					}
					return
				}
				w.WriteHeader(http.StatusOK)
				message := "Added commits successfully"
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
				return
			}
		default:
			{
				w.WriteHeader(http.StatusNotFound)
				message := "Invalid method"
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
				break
			}
		}
	}
}

func releaseCountHandler(collections *Collections) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			{
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
				w.Header().Set("Access-Control-Max-Age", "15")
				count, err := collections.Releases.CountDocuments(context.TODO(), bson.M{})
				if err == nil {
					res, err := json.Marshal(CommitCount{Count: count})
					if err == nil {
						w.WriteHeader(http.StatusOK)
						w.Write(res)
					} else {
						w.WriteHeader(http.StatusInternalServerError)
						message := "Could not create JSON for commit count data"
						status, err := json.Marshal(ResponseStatus{message})
						if err == nil {
							log.Println(message)
							w.Write(status)
						} else {
							writingStatusFailed(message)
						}
					}
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					message := "Could not retrieve number of releases"
					status, err := json.Marshal(ResponseStatus{message})
					if err == nil {
						log.Println(message)
						w.Write(status)
					} else {
						writingStatusFailed(message)
					}
				}
				break
			}
		default:
			{
				w.WriteHeader(http.StatusNotFound)
				message := "Invalid method"
				status, err := json.Marshal(ResponseStatus{message})
				if err == nil {
					log.Println(message)
					w.Write(status)
				} else {
					writingStatusFailed(message)
				}
			}
		}
	}
}
