package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	var err error
	if len(os.Args) < 2 {
		err = godotenv.Load(".env")
	} else {
		err = godotenv.Load(path.Join(os.Args[1], ".env"))
	}
	if err != nil {
		log.Fatalf("Error occured loading environment variables: %s", err)
	}
	r := mux.NewRouter()
	var collections Collections
	ConnectDB(&collections)
	dur := time.Duration(4 * time.Hour)
	periodicChannel := time.Tick(dur)
	go func() {
		for {
			<-periodicChannel
			res := collections.Config.FindOne(context.Background(), bson.M{})
			if res == nil {
				log.Fatalf("Could not retrieve server config")
			}
			var config ServerConfig
			err = res.Decode(&config)
			if err != nil {
				log.Fatalf("Error occured while decoding the server config")
			}
			wakatimeExpiry, err := time.Parse(time.RFC3339, config.Wakatime.ExpiresAt)
			if err != nil {
				diff := time.Until(wakatimeExpiry)
				if diff.Hours() < 24.0 {
					reqData := url.Values{
						"client_id":     {config.Wakatime.ClientID},
						"client_secret": {config.Wakatime.ClientSecret},
						"redirect_uri":  {"https://qat.dev"},
						"refresh_token": {config.Wakatime.RefreshToken},
						"grant_type":    {"refresh_token"},
					}
					resp, err := http.PostForm(config.Wakatime.RefreshURL, reqData)
					if err != nil {
						log.Fatalf("Error occured while refreshing the Wakatime token")
					}
					var bodyBytes []byte
					_, err = resp.Body.Read(bodyBytes)
					if err != nil {
						log.Fatalf("Error while reading the bytes of the response of refreshing the Wakatime token")
					}
					bodyStr := string(bodyBytes)
					if strings.Contains(bodyStr, "&") {
						vals := strings.Split(bodyStr, "&")
						for i := 0; i < len(vals); i++ {
							if strings.Contains(vals[i], "=") {
								itemSplit := strings.Split(vals[i], "=")
								switch itemSplit[0] {
								case "access_token":
									config.Wakatime.AccessToken = itemSplit[1]
								case "refresh_token":
									config.Wakatime.RefreshToken = itemSplit[1]
								case "expires_at":
									{
										expStr, err := url.QueryUnescape(itemSplit[1])
										if err != nil {
											log.Fatalf("Error while decoding token expiry date string")
										}
										config.Wakatime.ExpiresAt = expStr
									}
								}
							} else {
								log.Fatalf("Error parsing item at index %d in the response of refreshing the wakatime token", i)
							}
						}
						updateRes, err := collections.Config.UpdateOne(context.Background(), bson.M{},
							bson.M{"$set": bson.M{
								"wakatime.accessToken":  config.Wakatime.AccessToken,
								"wakatime.refreshToken": config.Wakatime.RefreshToken,
								"wakatime.expiresAt":    config.Wakatime.ExpiresAt}})
						if err != nil {
							log.Fatalf("Error while updating Wakatime configuration")
						}
						if updateRes.ModifiedCount != 1 {
							log.Fatalf("Updating Wakatime configuration failed")
						}
					} else {
						log.Fatalf("Error while parsing the response for refreshing Wakatime token")
					}
				}
			}

		}
	}()
	if len(os.Args) != 2 {
		os.RemoveAll(os.Getenv("COMPILE_DIR"))
	} else {
		os.RemoveAll(path.Join(os.Args[1], os.Getenv("COMPILE_DIR")))
	}
	r.HandleFunc("/compile", compileHandler)
	r.HandleFunc("/releases", releaseListHandler(&collections))
	r.HandleFunc("/downloadedRelease", downloadedReleaseHandler(&collections))
	r.HandleFunc("/newCommits", newCommitsHandler(&collections))
	r.HandleFunc("/latestCommit", latestCommitHandler(&collections))
	r.HandleFunc("/releaseCount", releaseCountHandler(&collections))
	err = http.ListenAndServe(os.Getenv("HOST")+":"+os.Getenv("PORT"), r)
	log.Println("Server connection failed")
	panic(err)
}
