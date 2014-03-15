//A simple tool for converting png, gif & jpg images into tiles useful for gameboy programming, with possible further applications.
//In the simplest case, tiles are to be read left to right then top to bottom, and each tile is to be individually converted into GBC 2-bit format.
//Future versions may include ability to write more than one tile together, e.g., so that each tile of a 16x16 image are writted one after the other.
//Color palettes are determined on a tile by tile basis. Colors are ordered by sum of RGB highest → lowest, with B > G > R being used for ties.
//Additional colors are set to black, and an error printed. With -t enabled the first color on each tile will be treated as transparency (color 0),
//overriding color sorting in this instance.
package main

import (flag "github.com/ogier/pflag"
	"os"
 	"image"
	"log"
	"fmt"
	"image/color"
 	_ "image/png"
 	_ "image/jpeg"
	_ "image/gif")

	//Organise settings flags
	var dim uint
	var dimX uint
	var dimY uint
	var offset uint
	var offsetX uint
	var offsetY uint
	var spacing uint
	var spacingX uint
	var spacingY uint
	var spriteMode bool
	var outFormat string

type sImage interface {
	image.Image
	SubImage(r image.Rectangle) image.Image
}

type palette [4]color.Color

var palettes []palette

func main() {
	var file *os.File
	var err error

	//flag.UintVar(&dim, "d", 8, "Square dimension of each tile. Use only for square. Non multiple-of-8 values may cause undefined behaviour.")
	//flag.UintVar(&dimX, "w", 8, "Width of each tile. Currently and possibly eternally unimplemented")
	//flag.UintVar(&dimY, "h", 8, "Height of each tile. Currently and possibly eternally unimplemented")

	//The whole dimensioning thing doesn't really work with GB format. It will
	//probably be removed completely in future updates.
	dimX, dimY, dim = 8, 8, 8
	flag.UintVarP(&offset, "offset", "o", 0, "Offset of the first tile from both the top and left edge")
	flag.UintVarP(&offsetX, "xoffset", "x", 0, "Horizontal offset of first tile from left")
	flag.UintVarP(&offsetY, "yoffset", "y", 0, "Vertical offset of first tile from top")
	flag.UintVarP(&spacing, "spacing", "s", 0, "Distance between tiles")
	flag.UintVarP(&spacingX, "xspacing", "X", 0, "Horizontal distance between tiles")
	flag.UintVarP(&spacingY, "yspacing", "Y", 0, "Vertical distance between tiles")
	flag.BoolVarP(&spriteMode, "spritemode", "t", false, "Sets first color in tile as transparency (color 0)")
	flag.StringVarP(&outFormat, "format", "f", "0x%X, ", "C Style format for output data (printed in a loop for each byte")
	flag.Parse();

	//if dimX, dimY are unset
	if dimX == dimY && dimY == 8 {
		//use the value from dim instead
		dimY = dim
		dimX = dim
	}
	if offsetX == offsetY && offsetY == 0 {
		offsetX = offset
		offsetY = offset
	}
	if spacingX == spacingY && spacingY == 0 {
		spacingX = spacing
		spacingY = spacing
	}

	//Program uses arg0 as read directory for input
	var fname = flag.Arg(0)

	//Default behaviour is read from stdin
	if fname == "-" || fname == "" {
		//read from standard input
		file = os.Stdin
	} else {
		//read from file
		file, err = os.Open(fname)
		if err != nil {
			log.Fatal(err)
		}
	}


	//decode file into image
	var outputData []byte
	tileset, format, err := image.Decode(file)
	if err == nil {
		log.Println(fname, "decoded from format", format)
	} else {
		log.Fatal(err)
	}
	tilesetSize := tileset.Bounds()

	//iterate through every tile fully contained within image
	for y := offsetY; y + dimY - 1 < uint(tilesetSize.Max.Y); y += dimY + spacingY {
		for x := offsetX; x + dimX - 1 < uint(tilesetSize.Max.X); x += dimX + spacingX {
			tile := tileset.(sImage).SubImage(image.Rect(int(x),int(y),int(x+dimX),int(y+dimY)))
			//Elipsis explodes the slice
			outputData = append(outputData, Encode(tile)...)
			//append slice to data
		}
	}
	//output data to file or stdOut
	var outFile *os.File
	if flag.Arg(1) == "-" || flag.Arg(1) == "" {
		outFile = os.Stdout
		log.Println("Outputting to stdout")
	} else {
		outFile, err = os.Create(flag.Arg(1))
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Outputting to file")
	}
	for i := 0; i < len(outputData); i++ {
		if i%16 == 0 && i != 0 {
			_, err = outFile.WriteString("\n")
		}
		_, err = fmt.Fprintf(outFile, outFormat, []byte{outputData[i]})
		if err != nil {
			log.Fatal(err)
		}
		//_, err = outFile.WriteString(delimiter)
		//if err != nil {
		//	log.Fatal(err)
		//}
	}
	//End the file with a newline. Programs like that.
	fmt.Fprintf(outFile, "\n")
	return
}



func Encode(tile image.Image) []byte {
	var tilePalette palette
	var tileColors []color.Color
	for i := range tilePalette {
		tilePalette[i] = color.Gray{0}
	}
	//Could we do better with a map?
	//var tilePaletteMap map[color.Color]byte
	var colCount byte = 0;
	size := tile.Bounds()
	//Not a huge fan of using the globals here but size.Max.Y-size.Min.Y is
	//hella messy, and we really don't have a case where dim != 8
	var rawData = make([]byte, dimX*dimY)
	var data = make([]byte, dimX*dimY/4)
	//list all colors. Drop any colors more than 4
	for y := size.Min.Y; y < size.Max.Y; y++ {
		for x:= size.Min.X; x < size.Max.X; x++ {
			color := tile.At(x,y)
			colorFound := false
			for i := 0; i<int(colCount); i++ {
				if color == tileColors[i] {
					colorFound = true
					break
				}
			}
			if !colorFound {
				tileColors = append(tileColors, color)
				colCount++
			}
			if colCount >= 4 {
				break
			}
		}
		if colCount >= 4 {
			break
		}
	}

	//sort colors (checking for -t)
	var paletteFound bool
	for i := range palettes {
		//compare current palette against all in palettes. Shouldn't
		//run at all if no palettes are defined
		if palettes[i].compare(tileColors) {
			//Order of already defined palette to be preserved
			tilePalette = palettes[i]
			paletteFound = true
			//we're done, let's move on
			break
		}
	}
	//if palette does not match an existing palette
	if !paletteFound {
		//sort the palette nicely
		tilePalette = sliceToPalette(tileColors)
		tilePalette = tilePalette.sort()
		//add new palette to set of image palettes
		palettes = append(palettes, tilePalette)
	}

	//create slice of color indices
	var pixelCount uint
	for y := size.Min.Y; y < size.Max.Y; y++ {
		for x:= size.Min.X; x < size.Max.X; x++ {
			var i byte
			for i = 0; i < 4; i++ {
				if tile.At(x,y) == tilePalette[i] {
					break
				}
			}
			rawData[pixelCount] = i
			pixelCount++
		}
	}
	//loop until you find your color
	//set the index in the slice
	//"Encode" into gameboy format
	//for each row
	for i := 0; i < int(dimX*dimY/8); i += 1 {
		//for each pixel in the row
		for n:= 0; n<8; n++ {
			//I hope this works
			//First byte is less significant bits of first row
			data[2*i] = ((rawData[8*i+n] & 1) << (7-byte(n))) | data[2*i]
			//Second byte is more significant bits of second row
			data[2*i+1] = ((rawData[8*i+n] & 2) >> 1 << (7-byte(n))) | data[2*i+1]
		}
	}
	return data
}

//sort() sorts the colors in a palette approximately from
//brightest to dimmest using a simple bubblesort
func (p palette) sort() palette{
	//Since it's such a small list, we're not checking if swapping is still
	//occurring, just sorting through to max time. More efficient sorting
	//would be nice but likely isn't worth the effort
	var min int = 0 //unnecessarily large type because casting is annoying
	//Spritemode leaves the first color identified where it is, no sorting
	if spriteMode {
		min = 1
	}
	for i := 0; i<4; i++ {
		for j := min; j<3-i; j++ {
			r0, g0, b0, _ := p[j].RGBA()
			r1, g1, b1, _ := p[j+1].RGBA()
			if r1 + g1 + b1 > r0 + g0 + b0 {
				p[j], p[j+1] = p[j+1], p[j]
			} else if r1 + g1 + b1 == r0 + g0 + b0 {
				if g1 + b1 > g0 + b0 {
					p[j], p[j+1] = p[j+1], p[j]
				} else if g1 + b1 == g0 + b0 {
					if b1 > b0 {
						p[j], p[j+1] = p[j+1], p[j]
					}
				}
			}
		}
	}
	return p
}

//Compares a palette to an array of colors to determine if all colors are in the palette.
func (a palette) compare(b []color.Color) bool{
	var match bool
	//select a color to check
	for i := range b {
		//innocent until proven guilty
		match = false
		//is it anywhere in b?
		for j := range a {
			if a[j] == b[i] {
				match = true
				break
			}
		}
		//No? Not the same palette
		if !match {
			return false
		}
	}
	//No match fails, success!
	return true
}

func sliceToPalette(a []color.Color) palette {
	var p palette
	for i := range a {
		if i >= 4 {
			break
		}
		p[i] = a[i]
	}
	return p
}
