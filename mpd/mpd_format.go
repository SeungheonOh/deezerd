package mpd

func StateName(state uint) string {
  switch state {
    case STATE_PAUSE:
    return "pause"
  case STATE_PLAY:
    return "play"
  case STATE_STOP:
    return "stop"
  }
  return "stop"
}
