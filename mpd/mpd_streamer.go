package mpd

import (
	"errors"
	"io"

	"github.com/godeezer/lib/deezer"
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
