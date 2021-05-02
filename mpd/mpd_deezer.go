package mpd

import (
	"io"
	"log"
	"net/http"
	"time"
  "errors"

	bufra "github.com/avvmoto/buf-readerat"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
	"github.com/godeezer/lib/deezer"
	"github.com/snabb/httpreaderat"
)

type DeezerSongStreamer struct {
  Drained chan struct{}
  Song deezer.Song
  Streamer beep.StreamSeekCloser
  Format beep.Format
}

func NewDeezerSongStreamer() *DeezerSongStreamer{
  return &DeezerSongStreamer {
    Drained: make(chan struct{}),
  }
}

func (d *DeezerSongStreamer) Load(song deezer.Song) error {
  d.Close()

  if song.ID == "" {
    return errors.New("invalid song id")
  }

  req, _ := http.NewRequest("GET", song.DownloadURL(deezer.MP3128), nil)
	htrdr, err := httpreaderat.New(nil, req, nil)
	if err != nil {
		log.Fatalln("error creating htrdr", err)
    return err
	}
	bhtrdr := bufra.NewBufReaderAt(htrdr, 1024*1024)
	reader, err := NewSongReadSeeker(io.NewSectionReader(bhtrdr, 0, int64(song.FilesizeMP3128)), song.ID)
	if err != nil {
		log.Fatalln("error creating srdr", err)
    return err
	}
	streamer, format, err := mp3.Decode(reader)
	if err != nil {
		log.Fatal(err)
    return err
	}

  d.Song = song
  d.Streamer = streamer
  d.Format = format

  return nil
}

func (d *DeezerSongStreamer) Stream(samples [][2]float64) (n int, ok bool) {
  filled := 0
  for filled < len(samples) {
    if d.Streamer == nil {
      for i := range(samples[filled:]) {
        samples[i][0] = 0
        samples[i][1] = 0
      }
      break
    }

    n, ok := d.Streamer.Stream(samples[filled:])

    if !ok {
      d.Close()
      d.Drained <- struct{}{}
      return len(samples), true
    }
    filled += n
  }
  return len(samples), true
}

func (d *DeezerSongStreamer) Err() error {
  return nil
}

func (d *DeezerSongStreamer) Len() int {
  return d.Streamer.Len()
}

func (d *DeezerSongStreamer) Position() int {
  return d.Streamer.Position()
}

func (d *DeezerSongStreamer) Seek(p int) error {
  return d.Streamer.Seek(p)
}

func (d *DeezerSongStreamer) Close() error {
  d.Song = deezer.Song{}
  if d.Streamer != nil {
    err := d.Streamer.Close()
    d.Streamer = nil
    return err
  }
  return nil
}

type DeezerPlayer struct {
  stop bool
  client *deezer.Client

  streamer *DeezerSongStreamer
  samplrate beep.SampleRate
  control *beep.Ctrl
  volume *effects.Volume
}

func NewDeezerPlayer(client *deezer.Client) *DeezerPlayer {
  player := &DeezerPlayer {
    stop: true,
    client: client,
    streamer: NewDeezerSongStreamer(),
    samplrate: beep.SampleRate(44100),
  }

  player.control = &beep.Ctrl{Streamer: player.streamer, Paused: false}
  player.volume = &effects.Volume {
    Streamer: player.control,
    Base: 10,
    Volume: 0,
    Silent: false,
  }
	speaker.Init(player.samplrate, player.samplrate.N(time.Second/10))
  return player
}

func (p *DeezerPlayer) Elapsed() uint {
  elapsed := p.streamer.Format.SampleRate.D(p.streamer.Position()).Round(time.Second)
  return uint(elapsed.Seconds())
}

func (p *DeezerPlayer) State() uint {
  if p.streamer.Song.ID == "" {
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
  p.streamer.Close()
}

func (p *DeezerPlayer) Seek(sec float64) {
  speaker.Lock()
  p.streamer.Seek(p.streamer.Format.SampleRate.N(time.Duration(sec) * time.Second))
  speaker.Unlock()
}

func (p *DeezerPlayer) Start() {
  p.streamer.Close()

	speaker.Play(p.volume)

}

const (
  STATE_PLAY = iota
  STATE_STOP   // When player is not null
  STATE_PAUSE  // When player is paused
)

type DeezerMpd struct {
  done chan struct{}
  client *deezer.Client

  player *DeezerPlayer
  Playlist *Playlist
}

func NewDeezerMpd(arl string) (*DeezerMpd, error) {
	client, err := deezer.NewClient(arl)
	if err != nil {
		log.Fatalln("Error creating client:", err)
	}

  player := NewDeezerPlayer(client)
  player.Start()

  pl := NewPlaylist("deezer buffer")

  /*
  plsongs := []string {
    "e1m1",
    "blinding lights",
    "Colonel Bogey March",
    "Waltz 2",
    "Budapest",
    "Was ist des Deutschen Vaterland",
  }

  for _, song := range(plsongs) {
    query, err := client.Search(song, "", "", 0, -1)
    if err != nil {
      log.Fatalln("query failed", err)
    }
    pl.Add(query.Songs.Data[0])
  }
  */

  mpd := &DeezerMpd {
    done: make(chan struct{}),
    client: client,
    player: player,
    Playlist: pl,
  }

    go func() {
    for {
      select {
      case <-mpd.player.streamer.Drained:
        mpd.Next()
        break
      case <-mpd.done:
        return
      }
    }
  }()

  return mpd, nil
}

func (m *DeezerMpd) SearchAndAdd(song string) {
  query, err := m.client.Search(song, "", "", 0, -1)
  if err != nil {
    log.Fatalln("query failed", err)
  }
  log.Println(song, len(query.Songs.Data))
  if len(query.Songs.Data) <= 0 {
    return
  }
  m.Playlist.Add(query.Songs.Data[0])
}

func (m *DeezerMpd) State() uint {
  return m.player.State()
}

func (m *DeezerMpd) PlayCurr() {
  m.Play(m.Playlist.SongPos)
}

func (m *DeezerMpd) Play(songpos int) {
  m.player.streamer.Load(m.Playlist.At(songpos))
  m.Playlist.SetPos(songpos);
}

func (m *DeezerMpd) Stop() {
  m.player.Stop()
}

func (m *DeezerMpd) Next() {
  m.player.streamer.Load(m.Playlist.Next())
}

func (m *DeezerMpd) Prev() {
  m.player.streamer.Load(m.Playlist.Prev())
}

func (m *DeezerMpd) Vol(v int) {
  // I aint no audio expert
  // Someone please take a look
  if v == 0 {
    m.player.volume.Silent = true
    return
  }
  m.player.volume.Silent = false
  m.player.volume.Volume = float64(v - 50) / 100
}

func (m *DeezerMpd) Close() {
  speaker.Clear()
  m.done <- struct{}{}
}
