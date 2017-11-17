package kaleidoscope

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
)

type Stream struct {
	Data chan string
	src  io.ReadCloser
}

func (s *Stream) read() {
	reader := bufio.NewReader(s.src)
	for {
		line, _ := reader.ReadBytes('\n')
		var msg Message
		json.NewDecoder(bytes.NewReader(line)).Decode(&msg)
		data, err := base64.StdEncoding.DecodeString(msg.Data)
		if err != nil {
			s.Data <- ""
		} else {
			s.Data <- string(data)
		}
	}
}

func (s *Stream) Close() {
	defer close(s.Data)
	s.src.Close()
	s.src = nil
}

func (s Stream) IsRunning() bool {
	return s.src != nil

}
