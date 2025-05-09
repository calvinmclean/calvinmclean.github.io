package sourdough

import (
	"fmt"
	"time"
)

//go:generate protoc -I=proto --go_out=proto proto/sourdough.proto

// Measurements:
// 1. uint8. 255 max (1 byte)
//    - Scaling factors: 1275 max with 5, 2550 with 10. Reduces granularity of measurement
//    - Conceptual scaling by changing units (g, kg)
// 2. uint16. 65,536 max (2 bytes)
// 3. uint8 + 3 bit (11 bits). This allows 255*5=1275 max with a few bits for granularity
// 	  - Unfortunately filesystems normally just allow data at the byte-level so this extra efficiency is lost
//      without using additional strategies for dealing with bit-level data
//
// Time:
// 1. Have a start date and then each row appends an offset in hours + minutes (4 bytes for base, +2 for each offset)
//    - This really depends on the use case. Are we always going to parse every row or do we sometimes random access?
//    - Using 8-bit gives us only 255 minutes. 16-bit is 65535 which is 45 days. I could save more by going to the bit-level
//    - If using uint16, then it's 2 bytes which only saves 2 bytes per data (although it is half)
// 2. Unix (int64, 8 bytes)
// 3. 2025+0-04-16 08:30 (0-255 / 1-12 / 1-31 / 0-23 / 0-59) = (8 bits / 4 bits / 5 bits / 5 bits / 6 bits) = (28 bits)
// 4. 2025+0-week-day 08:30 (0-255 / 1-52 / 1-7 / 0-23 / 0-59) = (8 bits / 6 bits / 3 bits / 5 bits / 6 bits) = (28 bits)
// 5. Mintutes since 2025-01-01 (525,960 minutes per year) (4,294,967,295 uint32 gives 8,165.9580481405 years) (4 bytes)
//    - Might as well do 1970-01-01 as the starting point since 55 years doesn't matter
//
// FlourType:
// 1. uint8. Simple iota enum stored as a single byte
// 2. 3 bits. If I only have up to 7 types, then 3 bits will work but unfortunately isn't very usable being less than a bit

var DefaultSerializer Serializer = NewSerializer(TimeModeCompact)

type Data struct {
	Time         time.Time
	StarterGrams uint8
	FlourGrams   uint8
	WaterGrams   uint8
	FlourType    FlourType
}

func (sd Data) String() string {
	return fmt.Sprintf(
		"%s: %3dg starter + %3d water + %3dg %s flour",
		sd.Time.Format(time.DateTime), sd.StarterGrams, sd.WaterGrams, sd.FlourGrams, sd.FlourType,
	)
}

func (s *Data) MarshalBinary() ([]byte, error) {
	return DefaultSerializer.Encode(*s), nil
}

func (s *Data) UnmarshalBinary(in []byte) error {
	DefaultSerializer.Decode(in, s)
	return nil
}

type FlourType uint8

const (
	FlourTypeUnknown FlourType = iota
	FlourTypeAllPurpose
	FlourTypeWholeWheat
	FlourTypeRye
	FlourTypeBread
)

func (ft FlourType) String() string {
	switch ft {
	case FlourTypeAllPurpose:
		return "All Purpose"
	case FlourTypeWholeWheat:
		return "Whole Wheat"
	case FlourTypeRye:
		return "Rye"
	case FlourTypeBread:
		return "Bread"
	default:
		return "Unknown/Invalid"
	}
}
