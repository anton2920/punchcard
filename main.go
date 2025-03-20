package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"golang.org/x/term"
)

const PunchcardHeader = `┌─────┬─┬──────────────────────────────────────────────────────────────────┬────────┐`

const PunchcardFirstLine = `│     │ │                                                                  │        │`

const PunchcardTwelvethLine = `│     │ │                                                                  │        │`

const PunchcardHR1 = `│     │ ├──────────────────────────────────────────────────────────────────┼────────┤`

const PunchcardEleventhLine = `│     │ │                  F O R T R A N  S T A T E M E N T                │        │`

const PunchcardHR2 = `├─────┼─┼──────────────────────────────────────────────────────────────────┼────────┤`

var PunchcardDigitalLines = []string{
	`│00000│0│000000000000000000000000000000000000000000000000000000000000000000│00000000│`,
	`│11111│1│111111111111111111111111111111111111111111111111111111111111111111│11111111│`,
	`│22222│2│222222222222222222222222222222222222222222222222222222222222222222│22222222│`,
	`│33333│3│333333333333333333333333333333333333333333333333333333333333333333│33333333│`,
	`│44444│4│444444444444444444444444444444444444444444444444444444444444444444│44444444│`,
	`│55555│5│555555555555555555555555555555555555555555555555555555555555555555│55555555│`,
	`│66666│6│666666666666666666666666666666666666666666666666666666666666666666│66666666│`,
	`│77777│7│777777777777777777777777777777777777777777777777777777777777777777│77777777│`,
	`│88888│8│888888888888888888888888888888888888888888888888888888888888888888│88888888│`,
	`│99999│9│999999999999999999999999999999999999999999999999999999999999999999│99999999│`,
}

const PunchcardFooter = `└─────┴─┴──────────────────────────────────────────────────────────────────┴────────┘`

const PunchcardVerticalBar = '│'

const PunchcardHole = '⌷'

const PunchcardInvalid = '▒'

var Alphabet = map[byte]int{
	'0':  (1 << 0),
	'1':  (1 << 1),
	'2':  (1 << 2),
	'3':  (1 << 3),
	'4':  (1 << 4),
	'5':  (1 << 5),
	'6':  (1 << 6),
	'7':  (1 << 7),
	'8':  (1 << 8),
	'9':  (1 << 9),
	'A':  ((1 << 12) | (1 << 1)),
	'B':  ((1 << 12) | (1 << 2)),
	'C':  ((1 << 12) | (1 << 3)),
	'D':  ((1 << 12) | (1 << 4)),
	'E':  ((1 << 12) | (1 << 5)),
	'F':  ((1 << 12) | (1 << 6)),
	'G':  ((1 << 12) | (1 << 7)),
	'H':  ((1 << 12) | (1 << 8)),
	'I':  ((1 << 12) | (1 << 9)),
	'J':  ((1 << 11) | (1 << 1)),
	'K':  ((1 << 11) | (1 << 2)),
	'L':  ((1 << 11) | (1 << 3)),
	'M':  ((1 << 11) | (1 << 4)),
	'N':  ((1 << 11) | (1 << 5)),
	'O':  ((1 << 11) | (1 << 6)),
	'P':  ((1 << 11) | (1 << 7)),
	'Q':  ((1 << 11) | (1 << 8)),
	'R':  ((1 << 11) | (1 << 9)),
	'S':  ((1 << 0) | (1 << 2)),
	'T':  ((1 << 0) | (1 << 3)),
	'U':  ((1 << 0) | (1 << 4)),
	'V':  ((1 << 0) | (1 << 5)),
	'W':  ((1 << 0) | (1 << 6)),
	'X':  ((1 << 0) | (1 << 7)),
	'Y':  ((1 << 0) | (1 << 8)),
	'Z':  ((1 << 0) | (1 << 9)),
	'&':  (1 << 12),
	'-':  (1 << 11),
	'/':  ((1 << 0) | (1 << 1)),
	':':  ((1 << 2) | (1 << 8)),
	'#':  ((1 << 3) | (1 << 8)),
	'@':  ((1 << 4) | (1 << 8)),
	'\'': ((1 << 5) | (1 << 8)),
	'=':  ((1 << 6) | (1 << 8)),
	'"':  ((1 << 7) | (1 << 8)),
	// '¢':  ((1 << 12) | (1 << 2) | (1 << 8)),
	'.': ((1 << 12) | (1 << 3) | (1 << 8)),
	'<': ((1 << 12) | (1 << 4) | (1 << 8)),
	'(': ((1 << 12) | (1 << 5) | (1 << 8)),
	'+': ((1 << 12) | (1 << 6) | (1 << 8)),
	'|': ((1 << 12) | (1 << 7) | (1 << 8)),
	'!': ((1 << 11) | (1 << 2) | (1 << 8)),
	'$': ((1 << 11) | (1 << 3) | (1 << 8)),
	'*': ((1 << 11) | (1 << 4) | (1 << 8)),
	')': ((1 << 11) | (1 << 5) | (1 << 8)),
	';': ((1 << 11) | (1 << 6) | (1 << 8)),
	// '¬':  ((1 << 11) | (1 << 7) | (1 << 8)),
	' ': 0,
	',': ((1 << 0) | (1 << 3) | (1 << 8)),
	'%': ((1 << 0) | (1 << 4) | (1 << 8)),
	'_': ((1 << 0) | (1 << 5) | (1 << 8)),
	'>': ((1 << 0) | (1 << 6) | (1 << 8)),
	'?': ((1 << 0) | (1 << 7) | (1 << 8)),
}

func PrintCardFirstLine(line string) {
	var sb strings.Builder

	var i int
	for _, r := range PunchcardFirstLine {
		if r == PunchcardVerticalBar {
			sb.WriteRune(r)
			continue
		}

		if i < len(line) {
			if _, ok := Alphabet[line[i]]; ok {
				sb.WriteByte(line[i])
			} else {
				sb.WriteRune(PunchcardInvalid)
				for (i < len(line)) && ((line[i] & 0x80) > 0) {
					i++
				}
			}
		} else {
			sb.WriteRune(r)
		}
		i++
	}

	sb.WriteString("\r\n")
	fmt.Print(sb.String())
}

func PrintCardPunchedLine(line string, cardLine string, search int) {
	var sb strings.Builder

	var i int
	for _, r := range cardLine {
		if r == PunchcardVerticalBar {
			sb.WriteRune(r)
			continue
		}

		if (i < len(line)) && ((Alphabet[line[i]] & search) == search) {
			sb.WriteRune(PunchcardHole)
		} else {
			sb.WriteRune(r)
		}
		i++
	}

	sb.WriteString("\r\n")
	fmt.Print(sb.String())
}

func PrintCardTwelvethLine(line string) {
	PrintCardPunchedLine(line, PunchcardTwelvethLine, 1<<12)
}

func PrintCardEleventhLine(line string) {
	PrintCardPunchedLine(line, PunchcardEleventhLine, 1<<11)
}

func PrintCardDigitalLines(line string) {
	for i := 0; i < len(PunchcardDigitalLines); i++ {
		PrintCardPunchedLine(line, PunchcardDigitalLines[i], 1<<i)
	}
}

func PrintLineOnCard(line string) {
	fmt.Printf(PunchcardHeader + "\r\n")
	PrintCardFirstLine(line)
	PrintCardTwelvethLine(line)
	fmt.Printf(PunchcardHR1 + "\r\n")
	PrintCardEleventhLine(line)
	fmt.Printf(PunchcardHR2 + "\r\n")
	PrintCardDigitalLines(line)
	fmt.Printf(PunchcardFooter + "\r\n")
}

func DisplayCard() {
	PrintLineOnCard("")
}

func Usage() {
	fmt.Fprintln(os.Stderr, "usage: punchcard [-p] [file]")
	os.Exit(1)
}

func PrintFile(args []string) error {
	var file *os.File

	switch len(args) {
	case 0:
		file = os.Stdin
	case 1:
		var err error

		file, err = os.Open(args[0])
		if err != nil {
			return fmt.Errorf("failed to open source file: %v", err)
		}
	default:
		Usage()
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		PrintLineOnCard(scanner.Text())
	}

	return nil
}

var (
	FileContents []byte
	LineOffsets  []int
)

func CalculateLineOffsets() {
	var pos int

	for {
		LineOffsets = append(LineOffsets, pos)

		lineFeed := bytes.IndexByte(FileContents[pos:], '\n')
		if lineFeed == -1 {
			break
		}

		pos += lineFeed + 1
	}
}

func DeleteLine(n int) {

}

func GetLine(buffer []byte, n int) int {
	if n < len(LineOffsets)-1 {
		return copy(buffer, FileContents[LineOffsets[n]:LineOffsets[n+1]-1])
	} else {
		return copy(buffer, FileContents[LineOffsets[n]:])
	}
}

func PrintLine(line []byte) {
	var buf bytes.Buffer

	DisplayCard()

	buf.WriteString(ESC + "[16A" + ESC + "[1C")
	for i := 0; i < len(line); i++ {
		if (i == 5) || (i == 6) || (i == 72) {
			buf.WriteString(ESC + "[1C")
		}

		buf.WriteByte(line[i])
		buf.WriteString(ESC + "[1B" + ESC + "[1D")

		if (Alphabet[line[i]] & (1 << 12)) == (1 << 12) {
			buf.WriteRune(PunchcardHole)
			buf.WriteString(ESC + "[1D")
		}
		buf.WriteString(ESC + "[1B")

		buf.WriteString(ESC + "[1B")

		if (Alphabet[line[i]] & (1 << 11)) == (1 << 11) {
			buf.WriteRune(PunchcardHole)
			buf.WriteString(ESC + "[1D")
		}
		buf.WriteString(ESC + "[1B")

		buf.WriteString(ESC + "[1B")

		for j := 0; j < len(PunchcardDigitalLines); j++ {
			if (Alphabet[line[i]] & (1 << j)) == (1 << j) {
				buf.WriteRune(PunchcardHole)
				buf.WriteString(ESC + "[1D")
			}
			buf.WriteString(ESC + "[1B")
		}

		buf.WriteString(ESC + "[1C")
		buf.WriteString(ESC + "[15A")
	}

	os.Stdout.Write(buf.Bytes())
}

func WriteLine(buffer []byte, n int) {
}

const (
	ESC = "\033"

	Backspace = 127
)

var (
	LeftArrow  = []byte{27, 91, 68}
	RightArrow = []byte{27, 91, 67}
)

/* TODO(anton2920): think about displaying cards stacked on each other. */
func EditFile(args []string) error {
	if len(args) != 1 {
		Usage()
	}
	file, err := os.OpenFile(args[0], os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	FileContents, err = io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read entire file: %v", err)
	}
	CalculateLineOffsets()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to switch terminal to RAW mode: %v", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		term.Restore(int(os.Stdin.Fd()), oldState)
		os.Exit(0)
	}()

	line := make([]byte, 80)
	pos := 0

	var lineIndex int
	pos = GetLine(line, lineIndex)
	PrintLine(line[:pos])

	var quit bool
	for !quit {

		buffer := make([]byte, 10)
		n, err := os.Stdin.Read(buffer)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %v", err)
		}
		buffer = buffer[:n]

		fmt.Printf("%v", buffer)

		if len(buffer) == 1 {
			switch buffer[0] {
			case 'q':
				quit = true
			case Backspace:
				pos = 0
			default:
				if len(buffer) <= len(line)-pos {
					pos += copy(line[pos:], buffer)
				}
			}
		} else if len(buffer) == 3 {
			if bytes.Equal(buffer, RightArrow) {
				lineIndex++
				if lineIndex >= len(LineOffsets) {
					lineIndex = len(LineOffsets) - 1
				}
			} else if bytes.Equal(buffer, LeftArrow) {
				/* Left arrow. */
				lineIndex--
				if lineIndex < 0 {
					lineIndex = 0
				}
			}
			pos = GetLine(line, lineIndex)
			PrintLine(line[:pos])
		}

	}

	return nil
}

func main() {
	var err error

	printFlag := flag.Bool("p", false, "print file instead of editing")
	flag.Parse()

	if *printFlag {
		err = PrintFile(flag.Args())
	} else {
		err = EditFile(flag.Args())
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}
}
