package sourdough

import (
	"encoding/binary"
	"time"
)

type TimeMode int

const (
	TimeModeUnix TimeMode = iota
	TimeModeUnixMinute
	TimeModeCompact
)

type Serializer struct {
	TimeMode TimeMode
	DataSize int64
}

func NewSerializer(timeMode TimeMode) Serializer {
	var dataSize int64
	switch timeMode {
	case TimeModeCompact:
		dataSize = 7
	case TimeModeUnix:
		dataSize = 12
	case TimeModeUnixMinute:
		dataSize = 8
	}

	return Serializer{timeMode, dataSize}
}

func (s Serializer) Encode(sd Data) []byte {
	buf := make([]byte, DefaultSerializer.DataSize)

	buf[0] = sd.StarterGrams
	buf[1] = sd.FlourGrams
	buf[2] = sd.WaterGrams

	switch s.TimeMode {
	case TimeModeCompact:
		copy(buf[3:], encodeCompactDateAndFlourType(sd))
	case TimeModeUnix:
		buf[3] = byte(sd.FlourType)
		binary.LittleEndian.PutUint64(buf[4:], uint64(sd.Time.Unix()))
	case TimeModeUnixMinute:
		buf[3] = byte(sd.FlourType)
		binary.LittleEndian.PutUint32(buf[4:], uint32(sd.Time.Unix()/60))
	}

	return buf
}

func (s Serializer) Decode(in []byte, sd *Data) {
	sd.StarterGrams = in[0]
	sd.FlourGrams = in[1]
	sd.WaterGrams = in[2]

	switch s.TimeMode {
	case TimeModeCompact:
		decodeCompactDateAndFlourType(in[3:], sd)
	case TimeModeUnix:
		sd.FlourType = FlourType(in[3])
		unixTime := binary.LittleEndian.Uint64(in[4:])
		sd.Time = time.Unix(int64(unixTime), 0).UTC()
	case TimeModeUnixMinute:
		sd.FlourType = FlourType(in[3])
		minutes := binary.LittleEndian.Uint32(in[4:])
		sd.Time = time.Unix(int64(minutes)*60, 0).UTC()
	}
}

// year offset (8) | month (4) | day (5) | hour (5) | minute (6)
func encodeCompactDateAndFlourType(sd Data) []byte {
	// Year is simply one byte/uint8
	year := uint8(sd.Time.Year() - 2025)

	// Month is the left four bits of the 2nd byte
	month := uint8(sd.Time.Month() << 4)

	// Day is 5 bits split across the 2nd and 3rd bytes
	day := sd.Time.Day()
	dayPart1 := (uint8(day) & 0b000_1111_0) >> 1
	dayPart2 := (uint8(day) & 0b0000000_1) << 7

	// Hour is 5 bits in the middle of the 3rd byte
	hour := uint8(sd.Time.Hour()) << 2

	// Minute is 6 bits split across the 3rd and 4th bytes
	minute := sd.Time.Minute()
	minutePart1 := (uint8(minute) & 0b00_11_0000) >> 4
	minutePart2 := (uint8(minute) & 0b0000_1111) << 4

	// FlourType is just the final 4 bits
	flourType := uint8(sd.FlourType) & 0b0000_1111

	return []byte{
		year,
		month | dayPart1,
		dayPart2 | hour | minutePart1,
		minutePart2 | flourType,
	}
}

func decodeCompactDateAndFlourType(data []byte, sd *Data) {
	year := int(data[0]) + 2025

	month := data[1] >> 4

	dayPart1 := (data[1] << 1) & 0b000_1111_0
	dayPart2 := (data[2] >> 7) & 0b0000000_1
	day := dayPart1 | dayPart2

	hour := (data[2] >> 2) & 0b000_11111

	minutePart1 := (data[2] << 4) & 0b00_11_0000
	minutePart2 := (data[3] >> 4) & 0b0000_1111
	minute := minutePart1 | minutePart2

	sd.Time = time.Date(year, time.Month(month), int(day), int(hour), int(minute), 0, 0, time.UTC)
	sd.FlourType = FlourType(data[3] & 0b0000_1111)
}
