package main

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/google/uuid"
)

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
				return
			}
			dir := path.Join(os.Getenv("COMPILE_DIR"), uniq.String())
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
