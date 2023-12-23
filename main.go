package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type FileProcessResult struct {
	FileName string
	ResultFileName string
	Success bool
	ErrorMessage string
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter filenames separated by spaces:")	
	input, err := reader.ReadString('\n')

	if err != nil {
		fmt.Println("An error occured while reading input. Please try again", err)
		return
	}

	input = strings.TrimSpace(input)
	files := strings.Split(input, " ")
	// files := []string{ "./inputs/bat.jpg" }

	fmt.Println("+--------------------------------------------+")
	fmt.Println("+ Please input the Number of operation below +")
	fmt.Println("+ 1 - Change image color to gray             +")
	fmt.Println("+ 2 - Change image color to new color        +")
	fmt.Println("+--------------------------------------------+")

	fmt.Print("Enter operation number:")
	opeInput, err := reader.ReadString('\n')
	opeInput = strings.TrimSpace(opeInput)

	if err != nil {
		fmt.Println("An error occured while reading input. Please try again", err)
		return
	}

	var targetColorInput, newColorInput string

	resultChannel := make(chan FileProcessResult, len(files))

	if opeInput == "2" {
		fmt.Print("Enter target color (hex):")
		targetColorInput, _ = reader.ReadString('\n')
		targetColorInput = strings.TrimSpace(targetColorInput)
		
		fmt.Print("Enter new color (hex):")
		newColorInput, _ = reader.ReadString('\n')
		newColorInput = strings.TrimSpace(newColorInput)
	}

	var wg sync.WaitGroup

	for _, filename := range files {
		wg.Add(1) 

		func(f string) {
			defer wg.Done()
			processFile(f, opeInput, targetColorInput, newColorInput, resultChannel)
		}(filename)
	}

	// Make sure to wait for all go routines to finish
	wg.Wait()
	close(resultChannel)

	fmt.Println("----------------------------------------------");

	for result := range resultChannel {
		if result.Success {
			fmt.Println("Successfully processed file:", result.FileName, "👉", result.ResultFileName)
		} else {
			fmt.Println("Failed to processing file:", result.FileName, "with error `", result.ErrorMessage, "`")
		}
	}
}

func processFile(filename string, opeInput string, targetColorInput string, newColorInput string, resultChannel chan<- FileProcessResult) {
	file, err := os.Open(filename)

	if(err != nil) {
		// log.Fatal("Failed to open file",filename)	
		resultChannel <- FileProcessResult{
			FileName: filename,
			Success: false,
			ErrorMessage: err.Error(),
		}

		return;
	}
	
	defer file.Close()
	
	img,fileType,err  := image.Decode(file)

	if(err != nil) {
		resultChannel <- FileProcessResult{
			FileName: filename,
			Success: false,
			ErrorMessage: err.Error(),
		}

		return;
	}

	var newImage image.Image

	if(opeInput == "1") {
		newImage = changeToGray(img)
	} 
	
	if (opeInput == "2") {
		targetColor, _ := hexToRGBA(targetColorInput)
		newColor, _ := hexToRGBA(newColorInput)
		newImage = changeColor(targetColor, newColor, img)
	}		

	err, newFileName := generateFile(newImage, fileType)

	if err != nil {
		resultChannel <- FileProcessResult{
			FileName: filename,
			Success: false,
			ErrorMessage: err.Error(),
		}
	}

	resultChannel <- FileProcessResult{
		FileName: filename,
		Success: true,
		ErrorMessage: "",
		ResultFileName: newFileName,
	}
}

func generateFile(newImage image.Image, fileType string) (error, string) {
	fileName := generateTimeStampFilename(fileType)
	outputFile, err := os.Create(fileName)

	if(err != nil) {
		// log.Fatal("Failed to create new file")
		return err, "";
	}

	defer outputFile.Close()

	jpeg.Encode(outputFile, newImage, nil)

	return nil, fileName 
}

func generateTimeStampFilename(fileExtension string) string {
	return fmt.Sprintf("./outputs/%d.%s", time.Now().UnixNano(), fileExtension)
}

func changeToGray(img image.Image) image.Image {
	bounds := img.Bounds()

	newImage := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			grayColor := color.GrayModel.Convert(originalColor)
			newImage.Set(x, y, grayColor)
		}
	}

	return newImage
}

func hexToRGBA(hex string) (color.RGBA, error) {
	var rgba color.RGBA

	if(hex[0] != '#' || len(hex) != 7) {
		return rgba, fmt.Errorf("invalid hex color format")
	}

	val, err := strconv.ParseUint(hex[1:], 16, 32)

	if(err != nil) {
		return rgba, fmt.Errorf("failed to parsed color")
	}

	// By performing these operations, the code extracts the individual red, green, and blue components 
	// from a single 32-bit integer representation of a color and sets the alpha channel to full opacity. 
	rgba.R = uint8(val >> 16)
    rgba.G = uint8((val >> 8) & 0xff)
    rgba.B = uint8(val & 0xff)
    rgba.A = 0xff // fully opaque

	return rgba, nil
}

func changeColor(colorToChange color.RGBA, newColor color.RGBA, img image.Image) image.Image {
	bounds := img.Bounds()
	newImage := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
		
			if (originalColor.A == colorToChange.A && originalColor.G == colorToChange.G && originalColor.B == colorToChange.B) {
				newImage.SetRGBA(x, y, newColor)
			} else {
				newImage.Set(x, y, originalColor)
			}
		}
	}

	return newImage
}
