package main

import (
    "fmt"
    "net"
    "os"
    "log"
    "bufio"

    "github.com/SeungheonOh/deezerd/mpd"
)

const (
    CONN_HOST = "localhost"
    CONN_PORT = "6600"
    CONN_TYPE = "tcp"
)

func loader(id int, jobs <-chan string, done chan<- struct{}, mpd *mpd.DeezerMpd) {
  for j := range(jobs) {
    mpd.SearchAndAdd(j)
    done <- struct{}{}
  }
}

func main() {
	arl := os.Getenv("ARL")
  if arl == "" {
		log.Fatalln("Missing $ARL")
	}

  if len(os.Args) < 2 {
    log.Fatalln("need playlist file")
  }
  plfile, err := os.Open(os.Args[1])
  if err != nil {
    log.Fatalln(err)
  }
  defer plfile.Close()

  mpd, err := mpd.NewDeezerMpd(arl)
  if err != nil {
    panic(err)
  }
  defer mpd.Close()

  var songs []string
  scanner := bufio.NewScanner(plfile)
  for scanner.Scan() {
    songs = append(songs, scanner.Text())
  }

  jobs := make(chan string, len(songs))
  done := make(chan struct{}, len(songs))
  for id := 0; id <= 2; id++ {
    go loader(id, jobs, done, mpd)
  }
  for _, song := range(songs) {
    jobs <- song
  }
  close(jobs)
  for _,_ = range(songs) {
    <-done
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
