package main

//How to cross compile?
//https://opensource.com/article/21/1/go-cross-compiling

//powershell
// $Env:GOOS = "linux"; $Env:GOARCH = "amd64"; go build
//cmd
/*
set GOARCH=amd64
set GOOS=linux
go build
*/
import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
)

type (

	// Timestamp is a helper for (un)marhalling time
	Timestamp time.Time

	// HookMessage is the message we receive from Alertmanager
	HookMessage struct {
		Version           string            `json:"version"`
		GroupKey          string            `json:"groupKey"`
		Status            string            `json:"status"`
		Receiver          string            `json:"receiver"`
		GroupLabels       map[string]string `json:"groupLabels"`
		CommonLabels      map[string]string `json:"commonLabels"`
		CommonAnnotations map[string]string `json:"commonAnnotations"`
		ExternalURL       string            `json:"externalURL"`
		Alerts            []Alert           `json:"alerts"`
	}

	Alert struct {
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		StartsAt    string            `json:"startsAt,omitempty"`
		EndsAt      string            `json:"EndsAt,omitempty"`
	}
)

func main() {
	app := fiber.New()
	prometheus := fiberprometheus.New("webhook-receiver-service")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)
	// Info writes logs in the color blue with "INFO: " as prefix
	var Info = log.New(os.Stdout, "\u001b[34mINFO: \u001B[0m", log.LstdFlags|log.Lshortfile)

	arg := flag.String("path", "./", "a string")
	flag.Parse()

	var path = strings.TrimSuffix(*arg, "/")
	fmt.Println("path:", path)
	// Warning writes logs in the color yellow with "WARNING: " as prefix
	//var Warning = log.New(os.Stdout, "\u001b[33mWARNING: \u001B[0m", log.LstdFlags|log.Lshortfile)

	// Error writes logs in the color red with "ERROR: " as prefix
	//var Error = log.New(os.Stdout, "\u001b[31mERROR: \u001b[0m", log.LstdFlags|log.Lshortfile)

	// Debug writes logs in the color cyan with "DEBUG: " as prefix
	//var Debug = log.New(os.Stdout, "\u001b[36mDEBUG: \u001B[0m", log.LstdFlags|log.Lshortfile)

	// GET /api/register
	app.Post("/sendMetrics", func(c *fiber.Ctx) error {
		var data HookMessage

		if err := c.BodyParser(&data); err != nil {
			Info.Println(err)
			return err
		}

		json, _ := json.MarshalIndent(data, "", " ")
		//fmt.Println(reflect.TypeOf(json))
		filePath := fmt.Sprintf("%s%s%d", path, "/", time.Now().UnixNano())
		//fmt.Println("path:", filePath)

		_ = ioutil.WriteFile(filePath, json, 0644)
		/*
			file, err := os.OpenFile(strconv.FormatInt(time.Now().UnixNano(), 10), os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				log.Println(err)
			}

			if _, err := file.WriteString(string(json[:])); err != nil {
				log.Fatal(err)
			}

			if _, err := file.WriteString("\n"); err != nil {
				log.Fatal(err)
			}
		*/
		//Info.Println(data)
		return c.SendStatus(fiber.StatusOK)
	})

	log.Fatal(app.Listen(":5001"))
}
