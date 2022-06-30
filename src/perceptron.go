/*
 * perceptron; a binary classifier.
 *
 * Henri Cattoire
 */
package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

type Perceptron struct {
	Threshold float64
	Weights   []float64
	Bounds    image.Rectangle
	P0        string
	P1        string
}

/*
 * train data (images)
 *
 * @param dataDir   name of the directory that contains the training data
 * @param p0        the perceptron doesn't fire for images that contain this substring
 * @param p1        the perceptron fires for images that contain this substring
 * @param epochs    amount of training cycles (can be less if the data converges early)
 *
 * @return a trained Perceptron model
 */
func Train(dataDir string, p0 string, p1 string, epochs int) Perceptron {
	var model Perceptron
	// TODO: add threshold option
	model.Threshold = 20.0
	model.P0 = p0
	model.P1 = p1

	reader, err := os.ReadDir(dataDir)
	ExitOnErr(err)
	if len(reader) > 0 {
		trainingSet := make([]string, len(reader))
		for i, file := range reader {
			trainingSet[i] = filepath.Join(dataDir, file.Name())
		}
		// FIXME: dataDir can only contain training data
		model.Weights, model.Bounds = InitWeights(trainingSet[0])
		// update weights until convergence or epochs
		correctClass := 0
		i := 0
		for correctClass != len(trainingSet) && i < epochs {
			correctClass = Cycle(trainingSet, model)
			i += 1
		}
		fmt.Println("perceptron: Error rate on trainingset:", (len(trainingSet)-correctClass)/len(trainingSet))
	} else {
		fmt.Fprintln(os.Stderr, "perceptron: empty training(set) directory!")
		os.Exit(1)
	}
	return model
}

/*
 * initialize weights to 0.0
 */
func InitWeights(sampleFileName string) ([]float64, image.Rectangle) {
	reader, err := os.Open(sampleFileName)
	ExitOnErr(err)
	defer reader.Close()
	im, _, err := image.Decode(reader)
	ExitOnErr(err)
	bounds := im.Bounds()
	return make([]float64, (bounds.Max.X-bounds.Min.X)*(bounds.Max.X-bounds.Min.X)), bounds
}

/*
 * perform one training cycle
 */
func Cycle(trainingSet []string, model Perceptron) int {
	var class string
	correctClass := 0
	for _, el := range trainingSet {
		class = Classify(el, model)
		if strings.Contains(filepath.Base(el), class) {
			correctClass += 1
		} else {
			// excitatory effect as p1 should be the response
			if class == model.P0 {
				UpdateWeights(el, 1, model)
			// inhibitory effect as p0 should be the response
			} else { // == model.p1
				UpdateWeights(el, -1, model)
			}
		}
	}
	return correctClass
}

/*
 * update model weights
 *
 * - constant > 0: make perceptron more likely to fire
 * - constant < 0: make perceptron less likely to fire
 */
func UpdateWeights(fileName string, constant float64, model Perceptron) {
	reader, err := os.Open(fileName)
	ExitOnErr(err)
	defer reader.Close()
	im, _, err := image.Decode(reader)
	ExitOnErr(err)
	var pixel color.Color
	bounds := im.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel = im.At(x, y)
			model.Weights[y*(bounds.Max.X-bounds.Min.X)+x] += constant * Luminosity(pixel)
		}
	}
}

/*
 * classify an image using the model
 */
func Classify(fileName string, model Perceptron) string {
	reader, err := os.Open(fileName)
	ExitOnErr(err)
	defer reader.Close()
	im, _, err := image.Decode(reader)
	ExitOnErr(err)
	if Response(im, model.Weights) > model.Threshold {
		return model.P1
	} else {
		return model.P0
	}
}

/*
 * calculate the response of the perceptron to an image
 */
func Response(im image.Image, weights []float64) float64 {
	var pixel color.Color
	response := 0.0
	bounds := im.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel = im.At(x, y)
			// TODO: make error prone (out of bounds?)
			response += Luminosity(pixel) * weights[y*(bounds.Max.X-bounds.Min.X)+x]
		}
	}
	return response
}

func Luminosity(color color.Color) float64 {
	R, G, B, _ := color.RGBA()
	return 0.299*float64(R) + 0.587*float64(G) + 0.114*float64(B) // Y
}

/*
 * io functions
 */

func SavePerceptron(perceptron Perceptron, fileName string) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(perceptron)
	ExitOnErr(err)
	file, err := os.Create(fileName)
	ExitOnErr(err)
	defer file.Close()
	buf.WriteTo(file)
}

func LoadPerceptron(fileName string) Perceptron {
	var buf bytes.Buffer
	var perceptron Perceptron
	dec := gob.NewDecoder(&buf)
	file, err := os.Open(fileName)
	ExitOnErr(err)
	defer file.Close()
	buf.ReadFrom(file)
	err = dec.Decode(&perceptron)
	ExitOnErr(err)
	return perceptron
}

func ToImage(perceptron Perceptron, fileName string) {
	// determine the range of the perceptron weights
	fromMin := perceptron.Weights[0]
	fromMax := perceptron.Weights[0]
	for _, pixel := range perceptron.Weights {
		if pixel < fromMin {
			fromMin = pixel
		}
		if pixel > fromMax {
			fromMax = pixel
		}
	}
	// create grayscale image
	bounds := perceptron.Bounds
	im := image.NewGray16(bounds)
	var weigth float64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			weigth = perceptron.Weights[y*(bounds.Max.X-bounds.Min.X)+x]
			im.SetGray16(x, y, color.Gray16{uint16(ToRange(weigth, fromMin, fromMax, 0, 65535))})
		}
	}
	file, err := os.Create(fileName)
	ExitOnErr(err)
	defer file.Close()
	png.Encode(file, im)
}

/*
 * convert x from inRange to outRange
 */
func ToRange(x, inMin, inMax, outMin, outMax float64) float64 {
	return (x-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}
