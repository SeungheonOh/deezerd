package mpd

import (
  "github.com/godeezer/lib/deezer"
)

type Playlist struct {
  Name string
  SongPos int
  queue []deezer.Song
}

func NewPlaylist(name string) *Playlist {
  queue := make([]deezer.Song, 0)
  return &Playlist {
    Name: name,
    SongPos: 0,
    queue: queue,
  }
}

func (p *Playlist) SetPos(pos int) {
  if pos >= len(p.queue) || pos < 0 {
    p.SongPos = 0
    return
  }
  p.SongPos = pos
}

func (p *Playlist) Curr() deezer.Song {
  if len(p.queue) == 0 || p.SongPos < 0 || p.SongPos > len(p.queue) {
    return deezer.Song{}
  }
  return p.queue[p.SongPos]
}

func (p *Playlist) Next() deezer.Song {
  p.SongPos += 1
  if p.SongPos >= len(p.queue) {
    p.SongPos = 0
  }
  return p.Curr()
}

func (p *Playlist) Prev() deezer.Song {
  p.SongPos -= 1
  if p.SongPos < 0 {
    p.SongPos = len(p.queue) - 1
  }
  return p.Curr()
}

func (p *Playlist) At(indx int) deezer.Song {
  if indx >= len(p.queue) || indx < 0 {
    return p.Curr()
  }
  return p.queue[indx]
}

func (p *Playlist) Add(songs ...deezer.Song) {
  p.queue = append(p.queue, songs...)
}

func (p *Playlist) Top() deezer.Song {
  return p.queue[0]
}

func (p *Playlist) Queue() []deezer.Song {
  return p.queue
}

func (p *Playlist) Pop() deezer.Song {
  defer func () {p.queue = p.queue[1:]}()
  return p.queue[0]
}
