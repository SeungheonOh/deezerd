package mpd

import (
	"io"
	"log"
	"net/http"
	"time"

	bufra "github.com/avvmoto/buf-readerat"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
	"github.com/godeezer/lib/deezer"
	"github.com/snabb/httpreaderat"
)

type DeezerPlayer struct {
  client *deezer.Client

  samplrate beep.SampleRate
  format *beep.Format
  streamer *beep.StreamSeekCloser
  control *beep.Ctrl
  volume *effects.Volume
}

func NewDeezerPlayer(client *deezer.Client) *DeezerPlayer {
  player := &DeezerPlayer {
    client: client,
    samplrate: beep.SampleRate(44100),
    volume : &effects.Volume {
      Base: 10,
      Volume: 0,
      Silent: false,
    },
  }

	speaker.Init(player.samplrate, player.samplrate.N(time.Second/10))
  return player
}

func (p *DeezerPlayer) Elapsed() uint {
  elapsed := (*p.format).SampleRate.D((*p.streamer).Position()).Round(time.Second)
  return uint(elapsed.Seconds())
}

func (p *DeezerPlayer) State() uint {
  if p.control == nil || p.streamer == nil {
    return STATE_STOP
  } else if p.control.Paused {
    return STATE_PAUSE
  } else {
    return STATE_PLAY
  }
}

func (p *DeezerPlayer) Pause(pause bool) {
  speaker.Lock()
  p.control.Paused = pause
  speaker.Unlock()
}

func (p *DeezerPlayer) Stop() {
  speaker.Clear()
  p.streamer = nil
  p.control = nil
}

func (p *DeezerPlayer) Seek(sec float64) {
  speaker.Lock()
  (*p.streamer).Seek((*p.format).SampleRate.N(time.Duration(sec) * time.Second))
  speaker.Unlock()
}

func (p *DeezerPlayer) Play(song deezer.Song) {
  p.Stop()
  req, _ := http.NewRequest("GET", song.DownloadURL(deezer.MP3128), nil)
	htrdr, err := httpreaderat.New(nil, req, nil)
	if err != nil {
		log.Fatalln("error creating htrdr", err)
	}
	bhtrdr := bufra.NewBufReaderAt(htrdr, 1024*1024)
	reader, err := NewSongReadSeeker(io.NewSectionReader(bhtrdr, 0, int64(song.FilesizeMP3128)), song.ID)
	if err != nil {
		log.Fatalln("error creating srdr", err)
	}
	streamer, format, err := mp3.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
  ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}
  p.volume.Streamer = ctrl
	speaker.Play(beep.Seq(p.volume, beep.Callback(func() {
		done <- true
	})))
  p.control = ctrl
  p.format = &format
  p.streamer = &streamer


  go func() {
    for {
      select {
      case <-done:
        p.Stop()
        return
      }
    }
  }()
}

const (
  STATE_PLAY = iota
  STATE_STOP   // When player is not null
  STATE_PAUSE  // When player is paused
)

type DeezerMpd struct {
  client *deezer.Client

  player *DeezerPlayer
  Playlist *Playlist
}

func NewDeezerMpd(arl string) (*DeezerMpd, error) {
	client, err := deezer.NewClient(arl)
	if err != nil {
		log.Fatalln("Error creating client:", err)
	}

  pl := NewPlaylist("deezer buffer")

  plsongs := []string {
    "e1m1",
    "blinding lights",
    "Colonel Bogey March",
    "Waltz 2",
  }

  for _, song := range(plsongs) {
    query, err := client.Search(song, "", "", 0, -1)
    if err != nil {
      log.Fatalln("query failed", err)
    }
    pl.Add(query.Songs.Data[0])

  }

  player := NewDeezerPlayer(client)
  return &DeezerMpd {
    client: client,
    player: player,
    Playlist: pl,
  }, nil
}

func (m *DeezerMpd) State() uint {
  return m.player.State()
}

func (m *DeezerMpd) PlayCurr() {
  m.player.Play(m.Playlist.Curr())
}

func (m *DeezerMpd) Play(songpos int) {
  m.player.Play(m.Playlist.At(songpos))
  m.Playlist.SetPos(songpos);
}

func (m *DeezerMpd) Next() {
  m.player.Play(m.Playlist.Next())
}

func (m *DeezerMpd) Prev() {
  m.player.Play(m.Playlist.Next())
}

func (m *DeezerMpd) Vol(v int) {
  // I aint no audio expert
  if v == 0 {
    m.player.volume.Silent = true
    return
  }
  m.player.volume.Silent = false
  m.player.volume.Volume = float64(v - 50) / 100
}
