package mpd

import (
  "io"
  "net"
  "bufio"
  "strings"
  "strconv"
  "errors"
  "fmt"
  "bytes"
)

type DeezerMpdServer struct {
  mpd *DeezerMpd
  conn net.Conn
  cmds map[string] func([]string) error

  ShouldRun bool
}

func NewDeezerMpdServer(c net.Conn, m *DeezerMpd) *DeezerMpdServer {
  cmds := make(map[string] func([]string) error)
  server := &DeezerMpdServer {
    mpd: m,
    conn: c,
    cmds: cmds,

    ShouldRun: true,
  }

  cmds["add"] = server.Add
  cmds["addid"] = server.Addid
  cmds["addtagid"] = server.Addtagid
  cmds["albumart"] = server.Albumart
  cmds["binarylimit"] = server.Binarylimit
  cmds["channels"] = server.Channels
  cmds["clear"] = server.Clear
  cmds["clearerror"] = server.Clearerror
  cmds["cleartagid"] = server.Cleartagid
  cmds["close"] = server.Close
  cmds["commands"] = server.Commands
  cmds["config"] = server.Config
  cmds["consume"] = server.Consume
  cmds["count"] = server.Count
  cmds["crossfade"] = server.Crossfade
  cmds["currentsong"] = server.Currentsong
  cmds["decoders"] = server.Decoders
  cmds["delete"] = server.Delete
  cmds["deleteid"] = server.Deleteid
  cmds["delpartition"] = server.Delpartition
  cmds["disableoutput"] = server.Disableoutput
  cmds["enableoutput"] = server.Enableoutput
  cmds["find"] = server.Find
  cmds["findadd"] = server.Findadd
  cmds["idle"] = server.Idle
  cmds["kill"] = server.Kill
  cmds["list"] = server.List
  cmds["listall"] = server.Listall
  cmds["listallinfo"] = server.Listallinfo
  cmds["listfiles"] = server.Listfiles
  cmds["listmounts"] = server.Listmounts
  cmds["listpartitions"] = server.Listpartitions
  cmds["listplaylist"] = server.Listplaylist
  cmds["listplaylistinfo"] = server.Listplaylistinfo
  cmds["load"] = server.Load
  cmds["lsinfo"] = server.Lsinfo
  cmds["mixrampdb"] = server.Mixrampdb
  cmds["mixrampdelay"] = server.Mixrampdelay
  cmds["mount"] = server.Mount
  cmds["move"] = server.Move
  cmds["moveid"] = server.Moveid
  cmds["moveoutput"] = server.Moveoutput
  cmds["newpartition"] = server.Newpartition
  cmds["next"] = server.Next
  cmds["notcommands"] = server.Notcommands
  cmds["outputs"] = server.Outputs
  cmds["outputset"] = server.Outputset
  cmds["partition"] = server.Partition
  cmds["password"] = server.Password
  cmds["pause"] = server.Pause
  cmds["ping"] = server.Ping
  cmds["play"] = server.Play
  cmds["playid"] = server.Playid
  cmds["playlist"] = server.Playlist
  cmds["playlistfind"] = server.Playlistfind
  cmds["playlistid"] = server.Playlistid
  cmds["playlistinfo"] = server.Playlistinfo
  cmds["playlistsearch"] = server.Playlistsearch
  cmds["plchanges"] = server.Plchanges
  cmds["plchangesposid"] = server.Plchangesposid
  cmds["previous"] = server.Previous
  cmds["prio"] = server.Prio
  cmds["prioid"] = server.Prioid
  cmds["random"] = server.Random
  cmds["rangeid"] = server.Rangeid
  cmds["readcomments"] = server.Readcomments
  cmds["readmessages"] = server.Readmessages
  cmds["readpicture"] = server.Readpicture
  cmds["repeat"] = server.Repeat
  cmds["replay_gain_mode"] = server.Replay_gain_mode
  cmds["replay_gain_status"] = server.Replay_gain_status
  cmds["rescan"] = server.Rescan
  cmds["search"] = server.Search
  cmds["searchadd"] = server.Searchadd
  cmds["searchaddpl"] = server.Searchaddpl
  cmds["seek"] = server.Seek
  cmds["seekcur"] = server.Seekcur
  cmds["seekid"] = server.Seekid
  cmds["sendmessage"] = server.Sendmessage
  cmds["setvol"] = server.Setvol
  cmds["shuffle"] = server.Shuffle
  cmds["single"] = server.Single
  cmds["stats"] = server.Stats
  cmds["status"] = server.Status
  cmds["stop"] = server.Stop
  cmds["subscribe"] = server.Subscribe
  cmds["swap"] = server.Swap
  cmds["swapid"] = server.Swapid
  cmds["tagtypes"] = server.Tagtypes
  cmds["toggleoutput"] = server.Toggleoutput
  cmds["unmount"] = server.Unmount
  cmds["unsubscribe"] = server.Unsubscribe
  cmds["update"] = server.Update
  cmds["urlhandlers"] = server.Urlhandlers
  cmds["volume"] = server.Volume

  return server
}

func (m *DeezerMpdServer) Handle() {
  m.conn.Write([]byte("OK MPD 0.24.2\n"));
  reader := bufio.NewReader(m.conn)
  for m.ShouldRun {
    rawreq, err := reader.ReadString('\n')
    if err != nil && err == io.EOF {
      break;
    }
    req := strings.TrimSpace(rawreq)

    if req != "" {
      fmt.Println(req)
    }

    args := strings.Split(strings.ReplaceAll(req, "\"", ""), " ")
    cmd, exists := m.cmds[args[0]]
    if exists {
      err := cmd(args[1:])
      if err == nil {
        m.conn.Write([]byte("OK\n"));
      }
    } else {
      fmt.Println("Command does not exists: ", req)
    }
  }
  fmt.Println("Closing ", m.conn.RemoteAddr().String())
  m.conn.Close()
}

func (m *DeezerMpdServer) Add(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Addid(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Addtagid(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Albumart(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Binarylimit(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Channels(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Clear(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Clearerror(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Cleartagid(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Close(args []string) error {
  m.ShouldRun = false
  return nil
}

// List all commands
func (m *DeezerMpdServer) Commands(args []string) error {
  var buffer bytes.Buffer
  for cmdskey := range m.cmds {
    fmt.Fprintf(&buffer, "command: %s\n", cmdskey)
  }
  m.conn.Write(buffer.Bytes())
  return nil
}

func (m *DeezerMpdServer) Config(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Consume(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Count(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Crossfade(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Currentsong(args []string) error {
  var buffer bytes.Buffer
  currsong := m.mpd.player.song
  fmt.Fprintf(&buffer, `title: %s
Album: %s
Artist: %s
Last-Modified: 2021-04-25T15:31:07Z
Time: %s
duration: %s.000
Pos: 1
Id: 1
`, currsong.Title, currsong.AlbumTitle, currsong.ArtistName, currsong.Duration, currsong.Duration)
  m.conn.Write([]byte(buffer.Bytes()))
  return nil
}

func (m *DeezerMpdServer) Decoders(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Delete(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Deleteid(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Delpartition(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Disableoutput(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Enableoutput(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Find(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Findadd(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Idle(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Kill(args []string) error {
  return nil
}

func (m *DeezerMpdServer) List(args []string) error {
  m.conn.Write([]byte(""))
  return nil
}

func (m *DeezerMpdServer) Listall(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Listallinfo(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Listfiles(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Listmounts(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Listpartitions(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Listplaylist(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Listplaylistinfo(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Load(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Lsinfo(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Mixrampdb(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Mixrampdelay(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Mount(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Move(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Moveid(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Moveoutput(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Newpartition(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Next(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Notcommands(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Outputs(args []string) error {
  m.conn.Write([]byte(`outputid: -
outputname: beep
plugin: beep
outputenabled: 1
`))
  return nil
}

func (m *DeezerMpdServer) Outputset(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Partition(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Password(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Pause(args []string) error {
  if len(args) == 0 {
    m.mpd.player.Pause(!m.mpd.player.control.Paused)
    return nil
  }

  if args[0] == "1" {
    m.mpd.player.Pause(true)
  } else {
    m.mpd.player.Pause(false)
  }
  return nil
}

func (m *DeezerMpdServer) Ping(args []string) error {
  m.conn.Write([]byte(""))
  return nil
}

func (m *DeezerMpdServer) Play(args []string) error {
  if len(args) == 0 {
    go m.mpd.player.Play(0)
    return nil
  }

  songpos, err := strconv.ParseInt(args[0], 10, 64)
  if err != nil {
    return err
  }
  go m.mpd.player.Play(int(songpos))
  return nil
}

func (m *DeezerMpdServer) Playid(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Playlist(args []string) error {
  m.conn.Write([]byte(""))
  return nil
}

func (m *DeezerMpdServer) Playlistfind(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Playlistid(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Playlistinfo(args []string) error {
  var buffer bytes.Buffer
  fmt.Fprintf(&buffer, `file: deezer stream
title: deezer stream
Album: 1aasdf
Artist: asf1
Duration: 123
Pos: 0
Id: 1
file: deezer stream
title: song2
Album: Braindead
Artist: hardcoding
Duration: 123
Pos: 1
Id: 2
`)//, m.mpd.player.song.Duration, m.mpd.player.song.Duration)
  m.conn.Write(buffer.Bytes())
  return nil
}

func (m *DeezerMpdServer) Playlistsearch(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Plchanges(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Plchangesposid(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Previous(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Prio(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Prioid(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Random(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Rangeid(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Readcomments(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Readmessages(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Readpicture(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Repeat(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Replay_gain_mode(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Replay_gain_status(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Rescan(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Search(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Searchadd(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Searchaddpl(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Seek(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Seekcur(args []string) error {
  if len(args) < 1 {
    return errors.New("not enough argument")
  }
  sec, err := strconv.ParseFloat(args[0], 64)
  if err != nil {
    return err
  }
  go m.mpd.player.Seek(sec)
  return nil
}

func (m *DeezerMpdServer) Seekid(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Sendmessage(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Setvol(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Shuffle(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Single(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Stats(args []string) error {
  return nil
}

// print status
func (m *DeezerMpdServer) Status(args []string) error {
  var buffer bytes.Buffer
  state := m.mpd.State()
  fmt.Fprintf(&buffer, `repeat: 0
random: 0
single: 0
consume: 0
partition: default
playlist: 2
playlistlength: 48
mixrampdb: 0.000000
volume: 100
state: %s
song: 9
songid: 1
`,
  StateName(state))

  if state != STATE_STOP {
    elapsed := m.mpd.player.Elapsed()
    fmt.Fprintf(&buffer, `time: %d:%s
elapsed: %d
bitrate: 0
duration: %s
audio: 44100:24:2
nextsong: 10
nextsongid: 11
`,
    elapsed,
    m.mpd.player.song.Duration,
    elapsed,
    m.mpd.player.song.Duration)
  }

  m.conn.Write(buffer.Bytes())
  return nil
}

func (m *DeezerMpdServer) Stop(args []string) error {
  m.mpd.player.Stop()
  return nil
}

func (m *DeezerMpdServer) Subscribe(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Swap(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Swapid(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Tagtypes(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Toggleoutput(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Unmount(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Unsubscribe(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Update(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Urlhandlers(args []string) error {
  return nil
}

func (m *DeezerMpdServer) Volume(args []string) error {
  return nil
}
