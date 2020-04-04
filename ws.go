package main

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	fmodChan = make(chan chan []byte)
)

func reader(ws *websocket.Conn, closed chan bool) {
	defer ws.Close()
	defer func() { closed <- true }()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func writer(ws *websocket.Conn, modified chan []byte, closed chan bool) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case fn := <-modified:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.TextMessage, fn); err != nil {
				return
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case <-closed:
			return
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	closed := make(chan bool)
	modified := make(chan []byte)

	fmodChan <- modified

	go writer(ws, modified, closed)
	reader(ws, closed)
}

func initWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// empty structs should occupy no memory
	workers := make(map[chan []byte]struct{})

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", filepath.Base(event.Name))

					fn := strings.TrimRight(filepath.Base(event.Name), filepath.Ext(event.Name))

					if fn == "" {
						continue
					}

					for c := range workers {
						select {
						case c <- []byte(fn):
							break
						default:
							// we can't check if it's actually closed
							// but assume it is if it blocks
							// TODO: fix that, because this will lead
							// to hard to diagnose bugs.
							delete(workers, c)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Fatalln(err)
				}
				log.Println("error:", err)
			case c := <-fmodChan:
				workers[c] = struct{}{}
			}
		}
	}()

	err = watcher.Add(*path)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
