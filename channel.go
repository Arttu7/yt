package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
)

func play(messages chan string, wg *sync.WaitGroup) {
	wg.Add(1)
	for msg := range messages {
		video, _ := exec.Command("youtube-dl", "-f", "mp4", "-g", msg).Output()
		fmt.Println(string(video))
		cmd := exec.Command("omxplayer", "-o", "local", fmt.Sprintf("%s", strings.TrimRight(string(video), "\n")))
		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Waiting for command to finish...")
		err = cmd.Wait()
		log.Printf("Command finished with error: %v", err)
	}
	wg.Done()
}
func main() {
	messages := make(chan string)
	var wg sync.WaitGroup
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		vid := "https://www.youtube.com/watch?v=" + req.URL.Query().Get("vid")
		messages <- vid
		io.WriteString(w, vid)

	}
	http.HandleFunc("/hello", helloHandler)
	go play(messages, &wg)
	log.Fatal(http.ListenAndServe(":8080", nil))
	wg.Wait()
}
