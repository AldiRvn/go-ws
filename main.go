package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	appPort := ":" + "53846"

	router := gin.Default()
	router.Use(gzip.Gzip(gzip.BestSpeed))
	router.Use(cors.Default())

	basePath := "/api/v1"
	apiV1 := router.Group(basePath)
	apiV1.GET("/", func(ctx *gin.Context) {
		wsServer := websocket.Server{Handler: websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()
			newParam := make(chan map[string]any)
			go func() { //? Waiting param body to available
				defer ws.Close()
				for {
					param := map[string]any{}
					if err := json.NewDecoder(ws).Decode(&param); err != nil {
						log.Println(err)
						return
					}
					newParam <- param
				}
			}()

			offset := 0
			param := map[string]any{}
			var action string
			for {
				select {
				case newParam := <-newParam:
					log.Println("NEW PARAM!")
					param = newParam //? Update value of param if new param available
					paramAction, _ := newParam["action"].(string)
					if paramAction != "" {
						action = paramAction
					}
				case <-ctx.Request.Context().Done():
					return
				default:
					if len(param) == 0 { //? When request connect is made, param will always empty
						continue
					}
				}

				switch action {
				case "start":
					offset = 0
					action = "resume"
				case "resume":
				case "pause":
					continue
				case "cancel":
					offset = 0
					continue
				case "close":
					return
				}
				if offset == 30 {
					return
				}

				offset += 1
				asJson, _ := json.Marshal(offset)
				_, _ = ws.Write(asJson)
				time.Sleep(1 * time.Second)
			}
		})}
		wsServer.ServeHTTP(ctx.Writer, ctx.Request)
	})

	log.Println("Running at", appPort)
	log.Fatalln(router.Run(appPort))
}
