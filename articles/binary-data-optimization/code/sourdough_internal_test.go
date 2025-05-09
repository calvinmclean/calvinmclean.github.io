package sourdough

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompactDateAndFlourType(t *testing.T) {
	//  year offset (8) | month (4) | day (5) | hour (5) | minute (6)
	data := []byte{
		// [0] all year offset
		0b0000_0000,
		// [1] month (4) | day (4)
		0b0100_1111,
		// [2] day (1) | hour (5) | minute (2)
		0b0010_0101,
		// [3] minute (4) | flour type (4)
		0b1110_0001,
	}

	sd := Data{}
	decodeCompactDateAndFlourType(data, &sd)
	back := encodeCompactDateAndFlourType(sd)

	assert.Equal(t, data, back)
}
