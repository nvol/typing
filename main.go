package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
)

const maxLineLength = 70

var (
	red       = color.New(color.FgWhite).Add(color.BgRed)
	grn       = color.New(color.FgHiGreen)
	beginWith = flag.Uint("b", 1, "the text number to begin with")
	needHelp  = flag.Bool("h", false, "this help")
)

func main() {
	flag.Parse()

	if *beginWith == 0 || *needHelp {
		flag.Usage()
		return
	}

	println("Press ESC to quit.")

	texts := loadTexts("typing.txt")
	println(len(texts), "texts loaded")

	if *beginWith > uint(len(texts)) {
		println("Text", *beginWith, "not found!\n")
		flag.Usage()
		return
	}

	if *beginWith > 1 {
		println("Beginning with the text number", *beginWith)
	}

	for _, text := range texts[*beginWith-1:] {
		if err := runTypingTest(
			text,
		); err != nil {
			println(err.Error())
			return
		}
	}

	println()
	println("That's all for now. Just press ESC to quit.")
	waitForEscOrError()
}

func runTypingTest(txtStr string) error {
	dur, mistakeCount, err := followText(txtStr)
	if err != nil {
		return err
	}

	println()
	println("-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-")

	durS := dur.Seconds()
	if durS == 0.0 {
		return errors.New("division by zero... it shouldn't be like this")
	}
	symbCount := len([]rune(txtStr))
	typingRate := int(float64(symbCount)*60.0/durS + 0.5)
	mistakeRatioPercents := int(float64(mistakeCount) * 100.0 / float64(symbCount))
	println("typing rate:", typingRate, "symbols per minute")
	print("mistakes: ", mistakeRatioPercents, "%\n")
	println("total symbols:", symbCount, " duration:", dur.String())
	return nil
}

func followText(txtStr string) (time.Duration, int, error) {
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	println()
	println(txtStr)

	lines := strings.Split(txtStr, "\n")
	txt := make([][]rune, len(lines))
	typed := make([][]rune, len(lines))
	for ix, line := range lines {
		txt[ix] = []rune(line)
		typed[ix] = make([]rune, 0, len(txt[ix]))
	}

	x, y := 0, 0
	println()

	isFirstSymbolTyped := false
	mistakeCount := 0
	var startDt time.Time
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		} else if key == keyboard.KeyBackspace {
			if x > 0 {
				x--
				typed[y] = typed[y][:len(typed[y])-1]
				retypeColoredLine(typed[y], txt[y])
				print(" ") // delete backspaced symbol
				retypeColoredLine(typed[y], txt[y])
				continue
			}
		} else if key == keyboard.KeySpace || key == keyboard.KeyEnter {
			if x == len(txt[y]) && string(typed[y]) == string(txt[y]) {
				y++
				x = 0
				println()

				// check y
				if y == len(txt) {
					return time.Since(startDt), mistakeCount, nil
				}

				continue
			}

			// just a space
			if key == keyboard.KeySpace {
				char = rune(' ')
			}
		} else if key == keyboard.KeyEsc {
			return 0, 0, errors.New("interrupted")
		}

		// print char
		if char != 0x00 && x < len(txt[y]) {
			// start timer by the first typed symbol
			if !isFirstSymbolTyped {
				startDt = time.Now()
				isFirstSymbolTyped = true
			}
			if char == txt[y][x] {
				grn.Print(string(char))
			} else {
				mistakeCount++
				red.Print(string(char))
			}
			typed[y] = append(typed[y], char)
			x++
		}
	}
}

func retypeColoredLine(txt, expected []rune) {
	print("\r")
	for ix, v := range txt {
		if v == expected[ix] {
			grn.Print(string(v))
		} else {
			red.Print(string(v))
		}
	}
}

func waitForEscOrError() {
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	for {
		_, key, err := keyboard.GetKey()
		if err != nil || key == keyboard.KeyEsc {
			return
		}
	}
}

func loadTexts(filename string) (texts []string) {
	texts = make([]string, 0, 128)

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	s := string(b)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\r", "")
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	texts = strings.Split(s, "\n\n")

	// wrap words for long lines
	for ix := range texts {
		texts[ix] = wrapWords(strings.TrimSpace(texts[ix]))
	}

	return
}

func wrapWords(v string) string {
	if len([]rune(v)) <= maxLineLength {
		return v
	}

	r := []rune(v)
	ret := ""

	for len(r) > maxLineLength {
		ix := maxLineLength
		for r[ix] != rune(' ') {
			ix--
			if ix < 0 {
				panic(errors.New(
					"cannot wrap words for [" + string(r[:(maxLineLength>>1)]) + "...]",
				))
			}
		}

		ret += string(r[:ix]) + "\n"
		r = r[ix+1:]
	}
	ret += string(r)
	return ret
}
