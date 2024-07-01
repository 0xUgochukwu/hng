package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Endpoint: [GET] <example.com>/api/hello?visitor_name="Mark" (where <example.com> is your server origin)
// {
//   "client_ip": "127.0.0.1", // The IP address of the c.ester
//   "location": "New York" // The city of the c.ester
//   "greeting": "Hello, Mark!, the temperature is 11 degrees Celcius in New York"
// }

func getClientIPByHeaders(c *gin.Context) (ip string, err error) {

	ipSlice := []string{}
	ipSlice = append(ipSlice, c.GetHeader("X-Forwarded-For"))
	ipSlice = append(ipSlice, c.GetHeader("x-forwarded-for"))
	ipSlice = append(ipSlice, c.GetHeader("X-FORWARDED-FOR"))

	for _, v := range ipSlice {
		if v != "" {
			return v, nil
		}
	}
	err = errors.New("error: Could not find clients IP address from the c.est Headers")
	return "", err

}

func getLocation(ip string) string {

	res, err := http.Get("http://ip-api.com/json/" + ip)
	if err != nil || res.StatusCode != 200 {
		fmt.Println("Error: Could not get location")
		return "Unknown"
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "Unknown"
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "Unknown"
	}
	city, ok := result["city"].(string)
	if !ok || city == "" {
		return "Unknown"
	}

	return city
}

func getTemperature(ip string) (int, error) {
	latlongURL := fmt.Sprintf("https://ipapi.co/%s/latlong/", ip)
	res, err := http.Get(latlongURL)
	if err != nil {
		return 0, fmt.Errorf("error getting latlong: %v", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading latlong response: %v", err)
	}

	latlong := strings.Split(string(body), ",")
	if len(latlong) != 2 {
		return 0, fmt.Errorf("invalid latlong response")
	}

	weatherURL := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=&q=%s,%s", latlong[0], latlong[1])
	print(weatherURL)
	res, err = http.Get(weatherURL)
	if err != nil {
		return 0, fmt.Errorf("error getting weather: %v", err)
	}
	defer res.Body.Close()

	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading weather response: %v", err)
	}

	type WeatherResponse struct {
		Current struct {
			TempC float64 `json:"temp_c"`
		} `json:"current"`
	}

	var weatherResp WeatherResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return 0, fmt.Errorf("error unmarshalling weather response: %v", err)
	}

	print(weatherResp.Current.TempC)
	return int(weatherResp.Current.TempC), nil
}

func main() {
	router := gin.Default()

	router.GET("/api/hello", func(c *gin.Context) {
		clientIP, err := getClientIPByHeaders(c)
		if err != nil {
			clientIP = c.ClientIP()
		}
		location := getLocation(clientIP)
		fmt.Println(clientIP)
		temp, err := getTemperature(clientIP)
		tempStr := ""
		if err != nil {
			tempStr = "Unknown"
		} else {
			tempStr = strconv.Itoa(temp)
		}
		greeting := "Hello, " + c.Query("visitor_name") + "! The temperature is " + tempStr + " degrees Celsius in " + location

		c.JSON(http.StatusOK, gin.H{
			"client_ip": clientIP,
			"location":  location,
			"greeting":  greeting,
		})
	})

	router.Run(":8080")
}
