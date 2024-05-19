package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
)

type Update struct {
	Value int
	Time  time.Time
}

var previous = Update{Value: 0, Time: time.Time{}}
var current = Update{Value: 0, Time: time.Time{}}

func main() {
	buf, _ := os.ReadFile("index.html")
	s := string(buf)
	tmpl, _ := template.New("index").Parse(s)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, current)
		fmt.Fprintf(w, "")
	})

	http.HandleFunc("/current", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"previous\":{\"time\": %d, \"value\": %d}, \"next\":{\"time\": %d, \"value\": %d}}", previous.Time.Unix(), previous.Value, current.Time.Unix(), current.Value)
	})

	fmt.Println("Starting value update function...")
	go update_counter()
	fmt.Println("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func update_counter() {
	client := &http.Client{}
	update(client)

	for {
		update(client)
	}
}

func update(client *http.Client) {
	resp, _ := client.Get("https://hub.docker.com/v2/namespaces/nginx/repositories/nginx-ingress/")
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		var data map[string]interface{}
		_ = json.Unmarshal([]byte(string(bodyString)), &data)
		if err != nil {
			fmt.Printf("could not unmarshal json: %s\n", err)
		}
		if current.Value != int(data["pull_count"].(float64)) {
			if current.Time.IsZero() {
				current.Time = time.Now()
				current.Value = int(data["pull_count"].(float64))
			} else if previous.Time.IsZero() {
				previous = current
				current.Time = time.Time{}
				current.Value = 0
			} else {
				previous.Time = current.Time
				previous.Value = current.Value
				current.Time = time.Now()
				current.Value = int(data["pull_count"].(float64))
			}
		}
	}
	// fmt.Printf("%s: %d\n", current.Time, current.Value)
	resp.Body.Close()
	time.Sleep(10 * time.Second)
}
