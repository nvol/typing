package main

import (
	"errors"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
)

var red = color.New(color.FgWhite).Add(color.BgRed)
var grn = color.New(color.FgHiGreen)

func main() {
	println("Press ESC to quit.")

	// 1st level
	if err := runTypingTest(
		"Что нужно для того, чтобы разогреть коромысло? Вот парочка советов: используйте силу мысли!\n" +
			"Не сдавайтесь, даже если кажется, что шансов 0... Просто поверьте в омлет. Ведь это возможно!\n" +
			"И вот ещё: не откладывайте на завтра то, что можно было бы вообще не делать. Иногда бодрит.",
	); err != nil {
		println(err.Error())
		return
	}

	// 2nd level
	if err := runTypingTest(
		"When I was working as a coachman at the post office, a shaggy geologist knocked on my door\n" +
			"and, looking at the map on the white wall, he grinned at me. He told me how Taiga was crying.\n" +
			"She's lonely without a man - they don't have a coachman at the post office.\n" +
			"So it's time for us to hit the road!",
	); err != nil {
		println(err.Error())
		return
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
