package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Data struct {
	RegionCode string `json:"region_code"`
}

type Location struct {
	Data []Data `json:"data"`
}

type Cases struct {
	State string `json:"state"`
	Count int64  `json:"count"`
}

type CaseCount struct {
	Document Cases `json:"document"`
}

func main() {
	e := echo.New()
	var latitude, longitude string
	clientOptions := options.Client().
		ApplyURI("mongodb+srv://mongodb:mongodb@cluster0.afw9j.mongodb.net/myFirstDatabase?retryWrites=true&w=majority")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Get("https://data.covid19india.org/v4/min/data.min.json")
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var cases map[string]map[string]map[string]interface{}
	err = json.Unmarshal(body, &cases)
	e.GET("/cases", func(c echo.Context) error {
		latitude = c.QueryParam("latitude")
		longitude = c.QueryParam("longitude")
		resp, err := http.Get(fmt.Sprintf("http://api.positionstack.com/v1/reverse?access_key=%s&query=%s,%s&limit=1", "d19766524242289be8f7414ae9e226fd", latitude, longitude))
		if err != nil {
			log.Fatalln(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		var location Location
		err = json.Unmarshal(body, &location)
		if err != nil {
			return err
		}
		regionCode := location.Data[0].RegionCode
		return c.String(http.StatusOK, fmt.Sprintf("Total cases = %v", getData(regionCode)))
	})
	e.Logger.Fatal(e.Start(":1323"))
}

func getData(region string) int64{
	url := "https://data.mongodb-api.com/app/data-wikgl/endpoint/data/beta/action/findOne"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf("{\"collection\":\"covid\",\"database\":\"inshorts\",\"dataSource\":\"Cluster0\",\"filter\": {\"state\": \"%s\" }}", region))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Access-Control-Request-Headers", "*")
	req.Header.Add("api-key", "FHp8fOCSa7OaqjA8mujS8K568JxxnoGgYYgusLBTIKU1y5duOVmWYveqrJldFbJn")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	var cases CaseCount
	_ = json.Unmarshal(body, &cases)
	return cases.Document.Count
}

