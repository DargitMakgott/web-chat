package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

type Message struct {
	Text string `json:"text"`
	Name string `json:"name"`
	Time string `json:"time"`
}

var feed = []Message{}

type Server struct {
	connections map[*websocket.Conn]bool
}

func newServer() *Server {
	return &Server{
		connections: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWSOrderBook(ws *websocket.Conn) {
	fmt.Println("New icoming connection from client to orderbook feed", ws.RemoteAddr())

	for {
		payload := fmt.Sprintf("orderbook data -> %d\n", time.Now().UnixNano())
		ws.Write([]byte(payload))
		time.Sleep(time.Second * 2)
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	fmt.Println("New incoming connection from client: ", ws.RemoteAddr())

	s.connections[ws] = true

	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error while reading the connection: ", err)
			continue
		}
		msg := buf[:n]

		s.broadcast(msg)
		feed = append(feed, Message{Text: string(msg), Name: ws.RemoteAddr().String(), Time: time.Now().GoString()})
	}
}

func (s *Server) broadcast(b []byte) {
	for ws := range s.connections {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				fmt.Println("Write error:, err")
			}
		}(ws)
	}
}

func main() {
	server := newServer()
	http.Handle("/ws", websocket.Handler(server.handleWS))
	http.Handle("/orderbookfeed", websocket.Handler(server.handleWSOrderBook))
	go http.ListenAndServe("192.168.0.24:3000", nil)

	router := gin.Default()
	router.GET("/history", getFeed)
	router.Run("192.168.0.24:8080")
}

func getFeed(c *gin.Context) {
	fmt.Println("Giving histiry", feed)
	c.IndentedJSON(http.StatusOK, feed)
}
