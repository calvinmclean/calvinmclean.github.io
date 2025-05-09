package sourdough_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"sourdough"
	"sourdough/proto/gen"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var data []sourdough.Data
var protoData gen.DataList

func init() {
	size := 10_000
	data = make([]sourdough.Data, size)
	protoData.Data = make([]*gen.Data, size)
	for i := range size {
		data[i].Time = time.Date(
			2025+rand.Intn(50),          // Year: 2025 + up to 49 years
			time.Month(rand.Intn(12)+1), // Month
			rand.Intn(28)+1,             // Day
			rand.Intn(24)+1,             // Hour
			rand.Intn(60)+1,             // Minute
			0, 0, time.UTC,
		)
		data[i].StarterGrams = uint8(rand.Intn(256))
		data[i].FlourGrams = uint8(rand.Intn(256))
		data[i].WaterGrams = uint8(rand.Intn(256))
		data[i].FlourType = sourdough.FlourType(rand.Intn(int(sourdough.FlourTypeBread + 1)))

		protoData.Data[i] = &gen.Data{
			StarterGrams: uint32(data[i].StarterGrams),
			FlourGrams:   uint32(data[i].FlourGrams),
			WaterGrams:   uint32(data[i].WaterGrams),
			FlourType:    gen.FlourType(data[i].FlourType),
			Time:         timestamppb.New(data[i].Time),
		}
	}
}

func BenchmarkJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		out, _ := json.Marshal(data)

		var sd []sourdough.Data
		_ = json.Unmarshal(out, &sd)

		assert.ElementsMatch(b, data, sd)

		cp := make([]byte, len(out))
		copy(cp, out)
	}
}

func BenchmarkBinaryUnixMinute(b *testing.B) {
	sourdough.DefaultSerializer = sourdough.NewSerializer(sourdough.TimeModeUnixMinute)
	var out bytes.Buffer
	enc := gob.NewEncoder(&out)
	_ = enc.Encode(data)

	dec := gob.NewDecoder(&out)
	var sd []sourdough.Data
	_ = dec.Decode(&sd)

	assert.ElementsMatch(b, data, sd)

	outBytes := out.Bytes()
	cp := make([]byte, len(outBytes))
	copy(cp, outBytes)
}

func BenchmarkBinaryUnix(b *testing.B) {
	sourdough.DefaultSerializer = sourdough.NewSerializer(sourdough.TimeModeUnix)
	var out bytes.Buffer
	enc := gob.NewEncoder(&out)
	_ = enc.Encode(data)

	dec := gob.NewDecoder(&out)
	var sd []sourdough.Data
	_ = dec.Decode(&sd)

	assert.ElementsMatch(b, data, sd)

	outBytes := out.Bytes()
	cp := make([]byte, len(outBytes))
	copy(cp, outBytes)
}

func BenchmarkBinaryCompact(b *testing.B) {
	sourdough.DefaultSerializer = sourdough.NewSerializer(sourdough.TimeModeCompact)
	var out bytes.Buffer
	enc := gob.NewEncoder(&out)
	_ = enc.Encode(data)

	dec := gob.NewDecoder(&out)
	var sd []sourdough.Data
	_ = dec.Decode(&sd)

	assert.ElementsMatch(b, data, sd)

	outBytes := out.Bytes()
	cp := make([]byte, len(outBytes))
	copy(cp, outBytes)
}

func TestManyIterations(t *testing.T) {
	t.Skip()
	n := 1000

	jsonStart := time.Now()
	for range n {
		out, _ := json.Marshal(data)
		var sd []sourdough.Data
		_ = json.Unmarshal(out, &sd)
	}
	fmt.Println("JSON:", time.Since(jsonStart))

	sourdough.DefaultSerializer = sourdough.NewSerializer(sourdough.TimeModeUnixMinute)
	binaryStart := time.Now()
	for range n {
		var out bytes.Buffer
		enc := gob.NewEncoder(&out)
		_ = enc.Encode(data)

		dec := gob.NewDecoder(&out)
		var sd []sourdough.Data
		_ = dec.Decode(&sd)
	}
	fmt.Println("Binary:", time.Since(binaryStart))
}

func BenchmarkProto(b *testing.B) {
	data, err := proto.Marshal(&protoData)
	if err != nil {
		log.Fatalf("Failed to encode: %v", err)
	}

	var sd gen.DataList
	if err := proto.Unmarshal(data, &sd); err != nil {
		log.Fatalf("Failed to decode: %v", err)
	}

	// The assertion fails because of other fields in the Data.
	// Note that disabling this part makes the benchmark faster than the others since it's not doing this assertion
	// If the assertion is disabled on others, they are faster
	// assert.ElementsMatch(b, protoData.Data, sd.Data)
}
