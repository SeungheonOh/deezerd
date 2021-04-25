package mpd

import (
	"errors"
//	"fmt"
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

/*
status
repeat: 0
random: 0
single: 0
consume: 0
partition: default
playlist: 1
playlistlength: 0
mixrampdb: 0.000000
state: stop
*/

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
  speaker.Lock()
  elapsed := (*p.format).SampleRate.D((*p.streamer).Position()).Round(time.Second)
  speaker.Unlock()
  return uint(elapsed.Seconds())
}

func (p *DeezerPlayer) Play() {
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
	defer streamer.Close()

	done := make(chan bool)
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
  ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: false}
	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		done <- true
	})))
  p.control = ctrl
  p.format = &format
  p.streamer = &streamer

	after := time.After(5 * time.Second)
  ticker := time.NewTicker(time.Second)
  defer ticker.Stop()
	for {
		select {
		case <-done:
      return
		case <-after:
			//streamer.Seek(format.SampleRate.N(100 * time.Second))
      break
    case <- ticker.C:
      //log.Print("toggled")
      //ctrl.Paused = !ctrl.Paused
      speaker.Lock()
      log.Print(format.SampleRate.D(streamer.Position()).Round(time.Second))
      speaker.Unlock()
      break
		}
	}
}

const (
  STATE_PLAY = iota
  STATE_STOP // what's the difference between stop and pause?
  STATE_PAUSE
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

  query, err := client.Search("Black in Black", "", "", 0, -1)
  if err != nil {
		log.Fatalln("query failed", err)
  }

  song := query.Songs.Data[0]
  player := NewDeezerPlayer(song, client)
  go player.Play()
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
