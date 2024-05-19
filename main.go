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

var previous = Update{Value: 0, Time: time.Now()}
var current = Update{Value: 0, Time: time.Now()}

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

	for {
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
			previous.Time = current.Time
			previous.Value = current.Value
			current.Time, err = time.Parse(time.RFC3339, data["last_updated"].(string))
			current.Value = int(data["pull_count"].(float64))
		}
		fmt.Printf("%s: %d\n", current.Time, current.Value)
		resp.Body.Close()
		time.Sleep(10 * time.Second)
	}
}
