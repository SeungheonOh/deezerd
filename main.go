package main

import (
    "fmt"
    "net"
    "os"
    "log"

    "github.com/SeungheonOh/deezerd/mpd"
)

const (
    CONN_HOST = "localhost"
    CONN_PORT = "6601"
    CONN_TYPE = "tcp"
)

func main() {
	arl := os.Getenv("ARL")
  if arl == "" {
		log.Fatalln("Missing $ARL")
	}
  mpd, err := mpd.NewDeezerMpd(arl)
  if err != nil {
    panic(err)
  }

  l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
  if err != nil {
    fmt.Println("Error listening:", err.Error())
    os.Exit(1)
  }

  defer l.Close()
  fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
  for {
    conn, err := l.Accept()
    if err != nil {
      fmt.Println("Error accepting: ", err.Error())
      os.Exit(1)
    }
    go handleRequest(conn, mpd)
  }
}

func handleRequest(conn net.Conn, deezermpd *mpd.DeezerMpd) {
  fmt.Println("Serving ", conn.RemoteAddr().String())
  server := mpd.NewDeezerMpdServer(conn, deezermpd)
  server.Handle()
}
