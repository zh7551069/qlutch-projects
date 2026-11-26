package airuntime

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestShortAudioLimitUsesContainerDuration(t *testing.T) {
	if exceedsShortAudioLimit(wavBytes(8, 300)) {
		t.Fatal("five minutes was rejected")
	}
	if !exceedsShortAudioLimit(wavBytes(8, 301)) {
		t.Fatal("longer than five minutes was accepted")
	}
	shortMP3 := make([]byte, 417)
	shortMP3[0], shortMP3[1], shortMP3[2] = 0xFF, 0xFB, 0x90
	duration, ok := audioDuration(shortMP3)
	if !ok || duration <= 0 || duration > shortAudioLimit {
		t.Fatalf("mp3 duration %s ok %v", duration, ok)
	}
	longMP4 := make([]byte, 24)
	copy(longMP4, "mvhd")
	binary.BigEndian.PutUint32(longMP4[16:20], 1)
	binary.BigEndian.PutUint32(longMP4[20:24], 301)
	if !exceedsShortAudioLimit(longMP4) {
		t.Fatal("mp4 longer than five minutes was accepted")
	}
	if exceedsShortAudioLimit([]byte{0x01, 0x02, 0x03, 0x04}) {
		t.Fatal("unrecognized audio was rejected by a guessed limit")
	}
}

func wavBytes(sampleRate, seconds int) []byte {
	dataLen := sampleRate * seconds
	buf := make([]byte, 44+dataLen)
	copy(buf[0:], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:], uint32(36+dataLen))
	copy(buf[8:], "WAVE")
	copy(buf[12:], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:], 16)
	binary.LittleEndian.PutUint16(buf[20:], 1)
	binary.LittleEndian.PutUint16(buf[22:], 1)
	binary.LittleEndian.PutUint32(buf[24:], uint32(sampleRate))
	binary.LittleEndian.PutUint32(buf[28:], uint32(sampleRate))
	binary.LittleEndian.PutUint16(buf[32:], 1)
	binary.LittleEndian.PutUint16(buf[34:], 8)
	copy(buf[36:], "data")
	binary.LittleEndian.PutUint32(buf[40:], uint32(dataLen))
	return buf
}

func TestAudioDurationMatchesRate(t *testing.T) {
	got, ok := audioDuration(wavBytes(8, 1))
	if !ok || got != time.Second {
		t.Fatalf("duration %s ok %v", got, ok)
	}
}
