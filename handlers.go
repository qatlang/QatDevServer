package main

import (
	"context"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
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
					statusData, err := json.Marshal(ReleasesListStatusFail{"Unable to find releases"})
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
				var status ReleasesListStatusFail
				status.Status = "Invalid request"
				statusBytes, err := json.Marshal(status)
				if err != nil {
					return
				}
				w.Write([]byte(statusBytes))
				break
			}
		}
	}
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
			uniq, err := uuid.NewUUID()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				status, err := json.Marshal(CompileStatusFail{"Cannot get UUID directory"})
				if err != nil {
					w.Write(status)
				}
				return
			}
			var dir string
			if len(os.Args) != 0 {
				dir = path.Join(os.Args[1], os.Getenv("COMPILE_DIR"), uniq.String())
			} else {
				dir = path.Join(os.Getenv("COMPILE_DIR"), uniq.String())
			}
			buildDir := path.Join(dir, "build")
			mainFile := path.Join(dir, "main.qat")
			err = os.MkdirAll(buildDir, fs.ModeDir)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				status, err := json.Marshal(CompileStatusFail{"Cannot create build directory"})
				if err != nil {
					w.Write(status)
				}
				os.RemoveAll(dir)
				return
			}
			err = os.WriteFile(mainFile, []byte(qatFile.Content), fs.ModeAppend)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				status, err := json.Marshal(CompileStatusFail{"Cannot write contents of file to compile"})
				if err != nil {
					w.Write(status)
				}
				os.RemoveAll(dir)
				return
			}
			cmd := exec.Command("qat", "run", mainFile, "-o", buildDir, "--no-colors")
			err = cmd.Run()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				status, err := json.Marshal(CompileStatusFail{"Running compiler failed: " + err.Error()})
				if err != nil {
					w.Write(status)
				}
				os.RemoveAll(dir)
				return
			}
			_, err = os.Stat(path.Join(dir, "build", "qat_result.json"))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				status, err := json.Marshal(CompileStatusFail{"Result file does not exist"})
				if err != nil {
					w.Write(status)
				}
				os.RemoveAll(dir)
				return
			}
			resContent, err := os.ReadFile(path.Join(dir, "build", "qat_result.json"))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				status, err := json.Marshal(CompileStatusFail{"Reading result file failed"})
				if err != nil {
					w.Write(status)
				}
				os.RemoveAll(dir)
				return
			}
			var sysCompRes SystemCompileResult
			err = json.Unmarshal(resContent, &sysCompRes)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				status, err := json.Marshal(CompileStatusFail{"Parsing result file failed"})
				if err != nil {
					w.Write(status)
				}
				os.RemoveAll(dir)
				return
			}
			if err == nil {
				w.Write(resContent)
				os.RemoveAll(dir)
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				status, err := json.Marshal(CompileStatusFail{"Converting result failed"})
				if err != nil {
					w.Write(status)
				}

			}
			break
		}
	default:
		{
			w.WriteHeader(http.StatusNotFound)
			break
		}
	}
}
