/*
 * perceptron; a binary classifier.
 *
 * Henri Cattoire
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	SHORT_USAGE = "Usage: perceptron <train|class> [ARGS]"
	// (sub)COMMANDS
	SHORT_TRAIN_USAGE = "Usage: perceptron train -0 <string> -1 <string> [ARGS] TRAININGSET"
	trainCmd          = flag.NewFlagSet("train", flag.ExitOnError)
	SHORT_CLASS_USAGE = "Usage: perceptron class [ARGS] IMAGE"
	classCmd          = flag.NewFlagSet("class", flag.ExitOnError)
	// Flags
	p0             string
	p1             string
	epochs         int
	modelFileName  string
	modelVisualize bool
)

func init() {
	trainCmd.StringVar(&p0, "0", "", "string to detect 0 perceived files.")
	trainCmd.StringVar(&p1, "1", "", "string to detect 1 perceived files.")
	trainCmd.IntVar(&epochs, "epochs", 1000, "number of training cycles.")
	trainCmd.BoolVar(&modelVisualize, "visualize", true, "visualize trained model as image (uses the name of the model).")
	trainCmd.StringVar(&modelFileName, "model", "perceptron.model", "file that stores the perceptron model.")
	classCmd.StringVar(&modelFileName, "model", "perceptron.model", "file that stores the perceptron model.")
}

/*
 * exit if error is not nil
 */
func ExitOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "perceptron:", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, SHORT_USAGE)
		os.Exit(1)
	}
	if os.Args[1] == "train" {
		trainCmd.Parse(os.Args[2:])
		if trainCmd.NArg() != 1 || p0 == "" || p1 == "" {
			fmt.Fprintln(os.Stderr, SHORT_TRAIN_USAGE)
			trainCmd.Usage()
			os.Exit(1)
		}
		model := Train(trainCmd.Arg(0), p0, p1, epochs)
		SavePerceptron(model, modelFileName)
		if modelVisualize {
			// TODO: make extension replacement more robust
			ToImage(model, strings.Replace(modelFileName, ".model", ".png", 1))
		}
	} else if os.Args[1] == "class" {
		classCmd.Parse(os.Args[2:])
		if classCmd.NArg() != 1 {
			fmt.Fprintln(os.Stderr, SHORT_CLASS_USAGE)
			classCmd.Usage()
			os.Exit(1)
		}
		model := LoadPerceptron(modelFileName)
		fmt.Println(Classify(classCmd.Arg(0), model))
	} else {
		fmt.Fprintln(os.Stderr, SHORT_USAGE)
		os.Exit(1)
	}
}
