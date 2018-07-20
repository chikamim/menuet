# Menuet
Golang library to create menubar apps- programs that live only in OSX's NSStatusBar

## Development Status

Under active development. API still changing rapidly.

## Installation
menuet requires OS X.

`go get github.com/caseymrm/menuet`

## Documentation

https://godoc.org/github.com/caseymrm/menuet

## Apps built with Menuet

* [Why Awake?](https://github.com/caseymrm/whyawake) - shows why your Mac can't sleep, and lets you force it awake
* [Not a Fan](https://github.com/caseymrm/notafan) - shows your Mac's temperature and fan speed, notifies you when your CPU is being throttled due to excessive heat

## Hello World

```go
package main

import (
	"time"

	"github.com/caseymrm/menuet"
)

func helloClock() {
	for {
		menuet.App().SetMenuState(&menuet.MenuState{
			Title: "Hello World " + time.Now().Format(":05"),
		})
		time.Sleep(time.Second)
	}
}
func main() {
	go helloClock()
	menuet.App().RunApplication()
}
```

![Output](https://github.com/caseymrm/menuet/raw/master/static/helloworld.gif)

## Advanced Features

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/caseymrm/menuet"
)

func temperature(woeid string) (temp, unit, text string) {
	url := "https://query.yahooapis.com/v1/public/yql?format=json&q=select%20item.condition%20from%20weather.forecast%20where%20woeid%20%3D%20" + woeid
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	var response struct {
		Query struct {
			Results struct {
				Channel struct {
					Item struct {
						Condition struct {
							Temp string `json:"temp"`
							Text string `json:"text"`
						} `json:"condition"`
					} `json:"item"`
					Units struct {
						Temperature string `json:"temperature"`
					} `json:"units"`
				} `json:"channel"`
			} `json:"results"`
		} `json:"query"`
	}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&response)
	if err != nil {
		log.Fatal(err)
	}
	return response.Query.Results.Channel.Item.Condition.Temp, response.Query.Results.Channel.Units.Temperature, response.Query.Results.Channel.Item.Condition.Text
}

var currentWoeid = "2442047"
var woeids = map[int]string{
	2442047: "Los Angeles",
	2487956: "San Francisco",
	2459115: "New York",
}

func setWeather() {
	temp, unit, text := temperature(currentWoeid)
	menuet.App().SetMenuState(&menuet.MenuState{
		Title: fmt.Sprintf("%s°%s and %s", temp, unit, text),
	})
}

func hourlyWeather() {
	for {
		setWeather()
		time.Sleep(time.Hour)
	}
}

func handleClick(woeid string) {
	currentWoeid = woeid
	setWeather()
}

func main() {
	go hourlyWeather()
	clickChannel := make(chan string)
	menuet.App().Label = "com.github.caseymrm.menuet.weather"
	menuet.App().Clicked = handleClick
	menuet.App().MenuOpened = func() []menuet.MenuItem {
		items := []menuet.MenuItem{}
		for woeid, name := range woeids {
			items = append(items, menuet.MenuItem{
				Text:     name,
				Callback: strconv.Itoa(woeid),
				State:    strconv.Itoa(woeid) == currentWoeid,
			})
		}
		return items
	}
	menuet.App().RunApplication()
}
```

![Output](https://github.com/caseymrm/menuet/raw/master/static/weather.png)

## License

MIT
