package mpd

import (
  "net"
  "bufio"
  "strings"
  "fmt"
)

type DeezerMpdServer struct {
  conn net.Conn
}

func NewDeezerMpdServer(c net.Conn) *DeezerMpdServer {
  return &DeezerMpdServer {
    conn: c,
  }
}

func (m *DeezerMpdServer) Handle() {
  for {
    reader := bufio.NewReader(m.conn)
    text, _ := reader.ReadString('\n')

    switch(strings.TrimSpace(text)) {
    case "STOP":
    case "END":
      goto CLOSE
    default:
      m.conn.Write([]byte("got message.\n"));
    }
  }
CLOSE:
  fmt.Println("Closing ", m.conn.RemoteAddr().String())
  m.conn.Write([]byte("bye.\n"));
  m.conn.Close()
}
