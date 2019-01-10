//Made by Arttu

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

var messages chan string

// Reads YT video ids from channel and then plays them using youtube-dl and omxplayer
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
func helloHandler(w http.ResponseWriter, req *http.Request) {
	//TODO: let users input also full YT addresses as well as just the id part; if vid is empty, skip playing
	url := req.URL.Query().Get("vid")
	if url != "" {
		parts := strings.Split(url, "https://youtu.be/")
		if len(parts) == 2 {
			url = parts[1]
		}
		vid := url
		startsWith := strings.HasPrefix(url, "https")
		if !startsWith {
			vid = "https://www.youtube.com/watch?v=" + url
		}
		messages <- vid
		io.WriteString(w, vid)
	} else {
		io.WriteString(w, "Video failed.")
	}
}

func main() {
	messages = make(chan string)
	var wg sync.WaitGroup
	http.HandleFunc("/hello", helloHandler)
	go play(messages, &wg)
	log.Fatal(http.ListenAndServe(":8080", nil))
	wg.Wait()
}
