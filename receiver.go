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
	"bytes"
	"encoding/csv"
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

	FMS struct {
		Identifier    string `json:"identifier"`
		Node          string `json:"node"`
		Suggestion    string `json:"suggestion"`
		AlarmName     string `json:"alarmname"`
		AlertGroup    string `json:"alertgroup"`
		Summary       string `json:"summary"`
		AlertKey      string `json:"alertkey"`
		Type          string `json:"type"` //1 and 2
		Severity      string `json:"severity"`
		Domain        string `json:"domain"`
		Manager       string `json:"manager"`
		AlarmPriority string `json:"alarmpriority"`
		RawSeverity   string `json:"rawseverity"`
	}

	NMS struct {
		Type   string  `json:"type"` //1 and 2
		Alerts []Alert `json:"alerts"`
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
		var fms FMS
		var nms NMS

		if err := c.BodyParser(&data); err != nil {
			Info.Println(err)
			return err
		}
		//log.Println(data.Status)
		//log.Println(data)
		//log.Println(data)
		//log.Println(data.Alerts[0])

		//fmt.Println(reflect.TypeOf(json))
		filePath := fmt.Sprintf("%s%s%d", path, "/", time.Now().UnixNano())

		//data_json, _ := json.MarshalIndent(data, "", " ")
		//_ = ioutil.WriteFile("./logs", data_json, 0644)
		//fmt.Println("path:", filePath)
		fms.Identifier = data.Status

		device_interface, ok := data.Alerts[0].Annotations["interface"]

		if !ok {
			device_interface = ""
		}

		fms.Identifier = data.Alerts[0].Labels["alertname"] + data.Alerts[0].Labels["function"] + data.Alerts[0].Labels["hostname"] + data.Alerts[0].Labels["type"] + device_interface
		fms.Node = data.Alerts[0].Labels["hostname"] + "-" + data.Alerts[0].Labels["instance"] + "-" + data.Alerts[0].Labels["function"]
		fms.AlarmName = data.Alerts[0].Labels["alertname"]
		fms.AlertGroup = data.Alerts[0].Labels["alertgroup"]
		fms.Summary = data.Alerts[0].Annotations["summary"]
		fms.Suggestion = data.Alerts[0].Annotations["suggestion"]
		fms.AlertKey = data.Alerts[0].Labels["function"]
		fms.Type = data.Status
		fms.Severity = data.Alerts[0].Labels["severity"]
		fms.Domain = data.Alerts[0].Labels["job"]
		fms.Manager = data.Alerts[0].Labels["devicetype"]
		fms.AlarmPriority = data.Alerts[0].Labels["alarmpriority"]

		pairs := [][]string{}
		pairs = append(pairs, []string{"identifier", "node", "alarmname", "alertgroup", "summary", "suggestion", "alertKey", "type", "severity", "domain", "manager", "alarmPriority"})
		pairs = append(pairs, []string{fms.Identifier, fms.Node, fms.AlarmName, fms.AlertGroup, fms.Summary, fms.Suggestion, fms.AlertKey, fms.Type, fms.Severity, fms.Domain, fms.Manager, fms.AlarmPriority})

		b := new(bytes.Buffer)
		w := csv.NewWriter(b)

		w.WriteAll(pairs)

		if err := w.Error(); err != nil {
			log.Fatal(err)
		}

		csvString := b.String()

		log.Println(csvString)

		//log.Println(fms.Node)
		nms.Alerts = data.Alerts
		nms.Type = data.Status

		// How to get value from map in golang
		//alertname, exists := data.CommonLabels["alertname"]
		//fmt.Printf("key exists in map: %t, value: %v \n", exists, alertname)
		/*
			nms_json, err := json.MarshalIndent(nms, "", " ")
			if err != nil {
				fmt.Println(err)
			}
			log.Println(nms)
			log.Println(nms_json)
		*/

		_ = ioutil.WriteFile(filePath+".csv", []byte(csvString), 0644)
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
