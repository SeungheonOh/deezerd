package mpd

import (
	"errors"
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

type SongReadSeeker struct {
	r        io.ReadSeeker
	cd       deezer.ChunkDecrypter
	cur      int64
	len      int64
	chunk    []byte
	chunkOff int64
}

func NewSongReadSeeker(r io.ReadSeeker, songid string) (*SongReadSeeker, error) {
	cd, err := deezer.NewChunkDecrypter(songid)
	if err != nil {
		return nil, err
	}
	len, err := r.Seek(0, io.SeekEnd)
	return &SongReadSeeker{
		r:     r,
		cd:    *cd,
		len:   len,
		chunk: make([]byte, 2048)}, nil
}

func (r *SongReadSeeker) Read(p []byte) (n int, err error) {
	for n < len(p) {
		chunkOff := r.cur / 2048
		var chunk []byte
		chunk, err = r.chunkAt(chunkOff)
		if err != nil {
			return
		}
		copied := copy(p[n:], chunk[r.cur%2048:])
		n += copied
		r.cur += int64(copied)
	}
  /*
	if err != nil {
		fmt.Println("read", n, err.Error())
	} else {
		fmt.Println("read", n)
	}
  */

	return
}

func (r *SongReadSeeker) Seek(off int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.cur = off
	case io.SeekCurrent:
		r.cur += off
	case io.SeekEnd:
		r.cur = r.len + off
	default:
		return 0, errors.New("invalid seek")
	}
	return r.cur, nil
}

func (r *SongReadSeeker) Close() error {
	return nil
}

func (r *SongReadSeeker) chunkAt(chunkoff int64) ([]byte, error) {
	if chunkoff == r.chunkOff {
		return r.chunk, nil
	}
	r.chunkOff = chunkoff
	pos, err := r.r.Seek(r.chunkOff*2048, io.SeekStart)
	if err != nil {
		return nil, err
	}
	n, err := io.ReadFull(r.r, r.chunk)
	if err != nil {
		if err == io.ErrUnexpectedEOF && pos+int64(n) == r.len {
			return r.chunk[:n], io.EOF
		}
		return nil, err
	}
	if r.chunkOff%3 == 0 {
		r.cd.DecryptChunk(r.chunk, r.chunk)
	}
	return r.chunk, nil
}

type DeezerPlayer struct {
  client *deezer.Client
  song deezer.Song

  format *beep.Format
  streamer *beep.StreamSeekCloser
  control *beep.Ctrl
  volume *effects.Volume
}

func NewDeezerPlayer(song deezer.Song, client *deezer.Client) *DeezerPlayer {
  return &DeezerPlayer {
    client: client,
    song: song,
  }
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

func (p *DeezerPlayer) Play(title int) {
  p.Stop()
  if title == 1 {
    query, err := p.client.Search("Colonel Bogey March", "", "", 0, -1)
    if err != nil {
      log.Fatalln("query failed", err)
    }

    p.song = query.Songs.Data[0]
  } else {
    query, err := p.client.Search("Blinding lights", "", "", 0, -1)
    if err != nil {
      log.Fatalln("query failed", err)
    }

    p.song = query.Songs.Data[0]

  }

  req, _ := http.NewRequest("GET", p.song.DownloadURL(deezer.MP3128), nil)
	htrdr, err := httpreaderat.New(nil, req, nil)
	if err != nil {
		log.Fatalln("error creating htrdr", err)
	}
	bhtrdr := bufra.NewBufReaderAt(htrdr, 1024*1024)
	reader, err := NewSongReadSeeker(io.NewSectionReader(bhtrdr, 0, int64(p.song.FilesizeMP3128)), p.song.ID)
	if err != nil {
		log.Fatalln("error creating srdr", err)
	}
	streamer, format, err := mp3.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
  ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}
	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		done <- true
	})))
  p.control = ctrl
  p.format = &format
  p.streamer = &streamer


  for {
    select {
    case <-done:
      p.Stop()
      return
    }
  }
}

const (
  STATE_PLAY = iota
  STATE_STOP   // When player is not null
  STATE_PAUSE  // When player is paused
)

type DeezerMpd struct {
  client *deezer.Client

  player *DeezerPlayer

  repeat bool
  state uint
  volume uint
}

func NewDeezerMpd(arl string) (*DeezerMpd, error) {
	client, err := deezer.NewClient(arl)
	if err != nil {
		log.Fatalln("Error creating client:", err)
	}

  query, err := client.Search("Blinding Lights", "", "", 0, -1)
  if err != nil {
		log.Fatalln("query failed", err)
  }

  song := query.Songs.Data[0]
  player := NewDeezerPlayer(song, client)
  return &DeezerMpd {
    player: player,
    repeat: false,
    state: STATE_STOP,
    volume: 100,
  }, nil
}

func (m *DeezerMpd) ChangeState(state uint) error {
  m.state = state
  return nil
}

func (m *DeezerMpd) State() uint {
  return m.player.State()
}
