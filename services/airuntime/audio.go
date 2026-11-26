package airuntime

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/bits"
	"time"
)

// shortAudioLimit is the synchronous recording boundary for this slice.
// Longer recordings are rejected. Asynchronous file transcription is not installed.
const shortAudioLimit = 300 * time.Second

func exceedsShortAudioLimit(body []byte) bool {
	duration, ok := audioDuration(body)
	if !ok {
		return false
	}
	return duration > shortAudioLimit
}

func audioDuration(body []byte) (time.Duration, bool) {
	if duration, ok := wavDuration(body); ok {
		return duration, true
	}
	if duration, ok := mp4Duration(body); ok {
		return duration, true
	}
	if duration, ok := mp3Duration(body); ok {
		return duration, true
	}
	return 0, false
}

func wavDuration(body []byte) (time.Duration, bool) {
	if len(body) < 12 || string(body[0:4]) != "RIFF" || string(body[8:12]) != "WAVE" {
		return 0, false
	}
	var (
		sampleRate uint32
		channels   uint16
		bitsPer    uint16
		dataBytes  uint64
		haveFmt    bool
		haveData   bool
	)
	off := 12
	for off+8 <= len(body) {
		id := string(body[off : off+4])
		size := int(binary.LittleEndian.Uint32(body[off+4 : off+8]))
		start := off + 8
		end := start + size
		if end > len(body) {
			end = len(body)
		}
		chunk := body[start:end]
		switch id {
		case "fmt ":
			if len(chunk) >= 16 {
				channels = binary.LittleEndian.Uint16(chunk[2:4])
				sampleRate = binary.LittleEndian.Uint32(chunk[4:8])
				bitsPer = binary.LittleEndian.Uint16(chunk[14:16])
				haveFmt = sampleRate > 0 && channels > 0 && bitsPer >= 8
			}
		case "data":
			dataBytes = uint64(len(chunk))
			haveData = true
		}
		off += 8 + size
		if size%2 == 1 {
			off++
		}
	}
	if !haveFmt || !haveData || bitsPer%8 != 0 {
		return 0, false
	}
	bytesPerSec := uint64(sampleRate) * uint64(channels) * uint64(bitsPer) / 8
	return durationFromRate(dataBytes, bytesPerSec)
}

func mp4Duration(body []byte) (time.Duration, bool) {
	idx := bytes.Index(body, []byte("mvhd"))
	if idx < 0 || idx+5 > len(body) {
		return 0, false
	}
	switch body[idx+4] {
	case 0:
		if idx+24 > len(body) {
			return 0, false
		}
		scale := binary.BigEndian.Uint32(body[idx+16 : idx+20])
		dur := binary.BigEndian.Uint32(body[idx+20 : idx+24])
		if scale == 0 {
			return 0, false
		}
		return durationFromRate(uint64(dur), uint64(scale))
	case 1:
		if idx+36 > len(body) {
			return 0, false
		}
		scale := binary.BigEndian.Uint32(body[idx+24 : idx+28])
		dur := binary.BigEndian.Uint64(body[idx+28 : idx+36])
		if scale == 0 {
			return 0, false
		}
		return durationFromRate(dur, uint64(scale))
	default:
		return 0, false
	}
}

func mp3Duration(body []byte) (time.Duration, bool) {
	payload := skipID3(body)
	if len(payload) < 4 {
		return 0, false
	}
	limit := len(payload) - 4
	if limit > 4096 {
		limit = 4096
	}
	for i := 0; i <= limit; i++ {
		bitrate, ok := mp3Bitrate(payload[i:])
		if !ok {
			continue
		}
		return durationFromRate(uint64(len(body)), uint64(bitrate/8))
	}
	return 0, false
}

func skipID3(body []byte) []byte {
	if len(body) < 10 || string(body[:3]) != "ID3" {
		return body
	}
	size := int(body[6]&0x7f)<<21 | int(body[7]&0x7f)<<14 | int(body[8]&0x7f)<<7 | int(body[9]&0x7f)
	off := 10 + size
	if off > len(body) {
		return nil
	}
	return body[off:]
}

func mp3Bitrate(header []byte) (int, bool) {
	if len(header) < 4 || header[0] != 0xFF || header[1]&0xE0 != 0xE0 {
		return 0, false
	}
	version := (header[1] >> 3) & 0x03
	layer := (header[1] >> 1) & 0x03
	bitrateIdx := (header[2] >> 4) & 0x0F
	sampleIdx := (header[2] >> 2) & 0x03
	if version == 0x01 || layer != 0x01 || bitrateIdx == 0 || bitrateIdx == 0x0F || sampleIdx == 0x03 {
		return 0, false
	}
	kbps := 0
	if version == 0x03 {
		table := [...]int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0}
		kbps = table[bitrateIdx]
	} else {
		table := [...]int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0}
		kbps = table[bitrateIdx]
	}
	if kbps == 0 {
		return 0, false
	}
	return kbps * 1000, true
}

func durationFromRate(units, perSecond uint64) (time.Duration, bool) {
	if perSecond == 0 {
		return 0, false
	}
	hi, lo := bits.Mul64(units, uint64(time.Second))
	if hi != 0 {
		return time.Duration(math.MaxInt64), true
	}
	return time.Duration(lo / perSecond), true
}

func imageMIME(body []byte) string {
	if len(body) >= 3 && body[0] == 0xFF && body[1] == 0xD8 && body[2] == 0xFF {
		return "image/jpeg"
	}
	if len(body) >= 8 && bytes.Equal(body[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		return "image/png"
	}
	if len(body) >= 12 && string(body[:4]) == "RIFF" && string(body[8:12]) == "WEBP" {
		return "image/webp"
	}
	if len(body) >= 12 && string(body[4:8]) == "ftyp" {
		switch string(body[8:12]) {
		case "heic", "heix", "hevc", "hevx", "mif1", "msf1":
			return "image/heic"
		case "heif":
			return "image/heif"
		}
	}
	return "application/octet-stream"
}

func audioMIME(body []byte) string {
	if len(body) >= 12 && string(body[:4]) == "RIFF" && string(body[8:12]) == "WAVE" {
		return "audio/wav"
	}
	if len(body) >= 12 && string(body[4:8]) == "ftyp" {
		return "audio/mp4"
	}
	if len(body) >= 3 && string(body[:3]) == "ID3" {
		return "audio/mpeg"
	}
	if len(body) >= 2 && body[0] == 0xFF && body[1]&0xE0 == 0xE0 {
		return "audio/mpeg"
	}
	return "application/octet-stream"
}
