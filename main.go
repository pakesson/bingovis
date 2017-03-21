package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
)

func usage(name string) {
	fmt.Println("Bin(Go)Vis")
	fmt.Println("")
	fmt.Printf("Usage: %v inputfile outputfile\n", name)
	fmt.Printf("   inputfile - Input file to analyze\n")
	fmt.Printf("   outputimage - Output PNG image\n")
}

func HilbertMapD2XY(n int, d int) (int, int) {
	x := 0
	y := 0

	rx := 0
	ry := 0
	t := d

	for s := 1; s < n; s *= 2 {
		rx = 1 & (t / 2)
		ry = 1 & (t ^ rx)

		if ry == 0 {
			if rx == 1 {
				x = s - 1 - x
				y = s - 1 - y
			}

			tmp := x
			x = y
			y = tmp
		}

		x += s * rx
		y += s * ry

		t /= 4
	}

	return x, y
}

func GetBlockCount(filelen int, blocklen int) int {
	x := float64(filelen) / float64(blocklen)
	for y := 0; ; y++ {
		val := math.Pow(math.Pow(2, float64(y)), 2)
		if val > x {
			return int(val)
		}
	}
	return 0
}

func GetBlock(data []byte, blockindex int, blocksize int) []byte {
	datalen := len(data)

	offset := blockindex * blocksize
	if offset >= datalen {
		return make([]byte, blocksize)
	}

	if offset+blocksize >= datalen {
		return data[offset:]
	}

	return data[offset : offset+blocksize]
}

func GetAverage(block []byte) uint8 {
	sum := 0
	for _, v := range block {
		sum += int(v)
	}
	return uint8(sum / len(block))
}

func GetEntropy(block []byte) uint8 {
	observations := make([]int, 256)
	var entropy float64 = 0

	for _, v := range block {
		observations[v] += 1
	}

	for i := 0; i < 256; i++ {
		probability := float64(observations[i]) / float64(len(block))
		if probability > 0 {
			entropy -= float64(probability) * math.Log2(float64(probability))
		}
	}

	return uint8(entropy * 32) // Change scaling?
}

func AnalyzeData(data []byte) (*image.RGBA, error) {
	datalen := len(data)
	if datalen == 0 {
		return nil, errors.New("Empty file")
	}

	fmt.Printf("Data: %d bytes\n", datalen)

	blocksize := 16
	blockcount := GetBlockCount(datalen, blocksize)
	sidelen := int(math.Sqrt(float64(blockcount)))

	fmt.Printf("Block size: %d\n", blocksize)
	fmt.Printf("Block count: %d\n", blockcount)
	fmt.Printf("Side length: %d\n", sidelen)

	img := image.NewRGBA(image.Rect(0, 0, sidelen, sidelen))

	for i := 0; i < blockcount; i++ {
		block := GetBlock(data, i, blocksize)

		red := GetEntropy(block)
		green := GetAverage(block)
		blue := uint8(0)

		x, y := HilbertMapD2XY(sidelen, i)

		img.Set(x, y, color.RGBA{red, green, blue, 255})
	}

	return img, nil
}

func GenerateBinVis(inputfile string, outputfile string) error {
	f, err := os.OpenFile(outputfile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	data, err := ioutil.ReadFile(inputfile)
	if err != nil {
		return err
	}

	img, err := AnalyzeData(data)
	if err != nil {
		return err
	}

	png.Encode(f, img)

	return nil
}

func main() {
	prog := filepath.Base(os.Args[0])

	if len(os.Args) != 3 {
		usage(prog)
		os.Exit(1)
	}

	inputfile := os.Args[1]
	outputfile := os.Args[2]

	err := GenerateBinVis(inputfile, outputfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		os.Exit(1)
	}
}
