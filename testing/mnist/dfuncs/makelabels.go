package dfuncs

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"os"
	"strconv"
)

const numLabels = 10
const pixelRange = 255

const (
	imageMagic = 0x00000803
	labelMagic = 0x00000801
	Width      = 28
	Height     = 28
)

type LabeledData struct {
	Data   []float32
	Number int
	Label  []float32
}

func (data LabeledData) MakeJPG(folder, name string, index int) error {
	dir := folder + "/" + name + "/"
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	newfile, err := os.Create(dir + strconv.Itoa(index) + ".jpg")
	if err != nil {
		return err
	}
	defer newfile.Close()
	image := data.convert()
	return jpeg.Encode(newfile, image, nil)
}
func (data LabeledData) MakeJPGsimple(index int) error {

	newfile, err := os.Create(strconv.Itoa(index) + ".jpg")
	if err != nil {
		return err
	}
	defer newfile.Close()
	image := data.convert()
	return jpeg.Encode(newfile, image, nil)
}

func (data LabeledData) convert() image.Image {
	var rect image.Rectangle

	rect.Min.X = 0
	rect.Min.Y = 0
	rect.Max.X = 28
	rect.Max.Y = 28
	img := image.NewRGBA(rect)
	for i := 0; i < 28; i++ {
		for j := 0; j < 28; j++ {
			pix := uint8(data.Data[i*28+j])

			img.Set(j, i, color.RGBA{pix, pix, pix, 255})
		}
	}
	return img
}
func LoadMNIST(filedirectory string, filenameLabel string, filenameData string) ([]LabeledData, error) {

	labelfile, err := os.Open(filedirectory + filenameLabel)
	if err != nil {
		//	panic(err)
		return nil, err
		//	panic(err)
	}
	alllabels, numbers, err := readLabelFile(labelfile)
	if err != nil {
		//	panic(err)
		return nil, err
	}
	datafile, err := os.Open(filedirectory + filenameData)
	if err != nil {
		//	panic(err)
		return nil, err
	}
	alldata, err := readImageFile(datafile)
	if err != nil {
		//	panic(err)
		return nil, err
	}
	if len(alldata) != len(alllabels) {
		return nil, errors.New("datafile and label file lengths don't match")
	}
	labeled := make([]LabeledData, len(alldata))
	for i := 0; i < len(alldata); i++ {
		labeled[i].Data = alldata[i]
		labeled[i].Label = alllabels[i]
		labeled[i].Number = numbers[i]
	}
	return labeled, nil
}

func readLabelFile(r io.Reader) ([][]float32, []int, error) {
	var err error

	var (
		magic int32
		n     int32
	)
	if err = binary.Read(r, binary.BigEndian, &magic); err != nil {
		//panic(err)
		return nil, nil, err
	}
	if magic != labelMagic {
		fmt.Println(magic, labelMagic)
		return nil, nil, os.ErrInvalid
	}
	if err = binary.Read(r, binary.BigEndian, &n); err != nil {
		return nil, nil, err
	}
	labels := make([][]float32, n)
	numbers := make([]int, n)
	for i := 0; i < int(n); i++ {
		var l uint8
		if err := binary.Read(r, binary.BigEndian, &l); err != nil {
			return nil, nil, err
		}
		numbers[i] = int(l)
		labels[i] = append(labels[i], makeonehotstate(l)...)
	}
	return labels, numbers, nil
}

//MakeEncodeeSoftmaxPerPixelCopy this is for NCHW
func MakeEncodeeSoftmaxPerPixelCopy(data []LabeledData) (copydata []LabeledData) {
	copydata = make([]LabeledData, len(data))
	for i, d := range data {
		copydata[i].Data = make([]float32, len(d.Data)*2)
		/*for i:=range d.Data{
			if d.Data[j]<128{
				copydata[i].d.Data[j]
			}
		}*/
		offset := len(d.Data)
		for j := range d.Data {
			if d.Data[j] < 128 {
				copydata[i].Data[j] = 0
				copydata[i].Data[offset+j] = 1
			} else {
				copydata[i].Data[j] = 1
				copydata[i].Data[offset+j] = 0
			}

		}
		copydata[i].Label = make([]float32, len(d.Label))
		copy(copydata[i].Label, d.Label)
		copydata[i].Number = d.Number
	}
	return copydata
}

func NormalizeData(data []LabeledData, average float32) []LabeledData {
	size := len(data)
	for i := 0; i < size; i++ {
		for j := 0; j < len(data[i].Data); j++ {

			data[i].Data[j] = (data[i].Data[j] - average) / float32(255)
		}

	}
	return data
}
func FindAverage(input []LabeledData) float32 {
	inputsize := len(input)
	datasize := len(input[0].Data)
	var adder float32
	for i := 0; i < inputsize; i++ {

		for j := 0; j < datasize; j++ {
			adder += input[i].Data[j]
		}
	}
	return adder / float32(inputsize*datasize)
}
func makeonehotstate(input uint8) []float32 {
	x := make([]float32, 10)
	x[input] = float32(1.0)
	return x

}
func readImageFile(r io.Reader) ([][]float32, error) {

	var err error

	var (
		magic int32
		n     int32
		nrow  int32
		ncol  int32
	)
	if err = binary.Read(r, binary.BigEndian, &magic); err != nil {
		return nil, err
	}
	if magic != imageMagic {
		return nil, err /*os.ErrInvalid*/
	}
	if err = binary.Read(r, binary.BigEndian, &n); err != nil {
		return nil, err
	}
	if err = binary.Read(r, binary.BigEndian, &nrow); err != nil {
		return nil, err
	}
	if err = binary.Read(r, binary.BigEndian, &ncol); err != nil {
		return nil, err
	}
	imgflts := make([][]float32, n)
	imgs := make([][]byte, n)
	size := int(nrow * ncol)
	for i := 0; i < int(n); i++ {
		imgflts[i] = make([]float32, size)
		imgs[i] = make([]byte, size)

		actual, err := io.ReadFull(r, imgs[i])
		if err != nil {
			return nil, err
		}
		if size != actual {
			return nil, os.ErrInvalid
		}
		for j := 0; j < size; j++ {
			imgflts[i][j] = float32(imgs[i][j])
		}
	}

	return imgflts, nil
}
