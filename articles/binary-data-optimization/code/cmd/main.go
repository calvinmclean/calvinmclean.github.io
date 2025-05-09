package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"sourdough"
)

func main() {
	size := 100
	data := make([]sourdough.Data, size)
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
	}

	filename := "data.bin"

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	enc := gob.NewEncoder(file)

	err = enc.Encode(data)
	if err != nil {
		log.Fatal(err)
	}

	// Reset to beginning of file
	_, err = file.Seek(0, 0)
	if err != nil {
		log.Fatal(err)
	}

	var newData []sourdough.Data
	dec := gob.NewDecoder(file)
	err = dec.Decode(&newData)
	if err != nil {
		log.Fatal(err)
	}

	for _, sd := range newData {
		fmt.Println(sd)
	}
}
