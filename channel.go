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
var pipe io.WriteCloser
var msg chan string
var skipall bool
var count int
var yt chan string

// Reads YT video ids from channel and then plays them using youtube-dl and omxplayer
func youtube(messages chan string) {
	for msg := range messages {
		count = count - 1
		if skipall {
			if count == 0 {
				skipall = false
			}
			continue
		}
		video, _ := exec.Command("youtube-dl", "-f", "mp4", "-g", msg).Output()
		fmt.Println(string(video))
		yt <- fmt.Sprintf("%s", strings.TrimRight(string(video), "\n"))
	}
}
func play(yt chan string, wg *sync.WaitGroup) {
	for video := range yt {
		if skipall {
			if count == 0 {
				skipall = false
			}
			continue
		}
		cmd := exec.Command("omxplayer", "-o", "local", video)
		pipe, _ = cmd.StdinPipe()
		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Waiting for command to finish...")
		err = cmd.Wait()
		log.Printf("Command finished with error: %v", err)
		pipe = nil
	}
	log.Println("Shutting down...")
	wg.Done()
}

func addHandler(w http.ResponseWriter, req *http.Request) {
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
		go addtochan(vid)
		io.WriteString(w, "Added to queue.\n")
	} else {
		io.WriteString(w, "Video failed.")
	}
}
func addtochan(vid string) {
	count = count + 1
	messages <- vid
}

func skipallHandler(w http.ResponseWriter, req *http.Request) {
	if pipe != nil {
		pipe.Write([]byte("q"))
		skipall = true
		io.WriteString(w, "All videos skipped.")
	}
}

func skipHandler(w http.ResponseWriter, req *http.Request) {
	if pipe != nil {
		pipe.Write([]byte("q"))
		io.WriteString(w, "Video skipped.")
	}

}
func closeHandler(w http.ResponseWriter, req *http.Request) {
	close(messages)
	if pipe != nil {
		pipe.Write([]byte("c"))
		io.WriteString(w, "Videoplayer stopped.")
	}

}
func pauseHandler(w http.ResponseWriter, req *http.Request) {
	if pipe != nil {
		pipe.Write([]byte("p"))
		io.WriteString(w, "Video paused.")
	}

}
func backwardHandler(w http.ResponseWriter, req *http.Request) {
	if pipe != nil {
		pipe.Write([]byte("\x1b[D"))
		io.WriteString(w, "Backed.")
	}

}
func forwardHandler(w http.ResponseWriter, req *http.Request) {
	if pipe != nil {
		pipe.Write([]byte("\x1b[C"))
		io.WriteString(w, "Backed.")
	}

}

func main() {
	messages = make(chan string)
	yt = make(chan string)
	var wg sync.WaitGroup
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/skip", skipHandler)
	http.HandleFunc("/skipa", skipallHandler)
	http.HandleFunc("/close", closeHandler)
	http.HandleFunc("/p", pauseHandler)
	http.HandleFunc("/b", backwardHandler)
	http.HandleFunc("/f", forwardHandler)
	log.Println("Welcome to youtube player!")
	wg.Add(1)
	go youtube(messages)
	go play(yt, &wg)
	go http.ListenAndServe(":8080", nil)
	wg.Wait()
}
