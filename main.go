package main

import (
	"bufio"
	stdbytes "bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	stdstrings "strings"

	"github.com/anton2920/gofa/bytes"
	"github.com/anton2920/gofa/ints"
	"github.com/anton2920/gofa/strings"

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
	var sb stdstrings.Builder

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
	var sb stdstrings.Builder

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

func MoveCursorUp(n int) string {
	return fmt.Sprintf(ESC+"[%dA", n)
}

func MoveCursorDown(n int) string {
	return fmt.Sprintf(ESC+"[%dB", n)
}

func MoveCursorRight(n int) string {
	return fmt.Sprintf(ESC+"[%dC", n)
}

func MoveCursorLeft(n int) string {
	return fmt.Sprintf(ESC+"[%dD", n)
}

func MoveCursorCol(n int) string {
	return fmt.Sprintf(ESC+"[%dG", n)
}

var SkipPositions = [...]int{5, 6, 72}

func WriteChar(w io.Writer, char byte, pos int) {
	for i := 0; i < len(SkipPositions); i++ {
		if pos == SkipPositions[i] {
			io.WriteString(w, MoveCursorRight(1))
			break
		}
	}

	if _, ok := Alphabet[char]; ok {
		w.Write([]byte{char})
	} else {
		io.WriteString(w, string(PunchcardInvalid))
	}
	io.WriteString(w, MoveCursorLeft(1))
	io.WriteString(w, MoveCursorDown(1))

	if (Alphabet[char] & (1 << 12)) == (1 << 12) {
		io.WriteString(w, string(PunchcardHole))
		io.WriteString(w, MoveCursorLeft(1))
	}
	io.WriteString(w, MoveCursorDown(1))

	/* Skip frame border. */
	io.WriteString(w, MoveCursorDown(1))

	if (Alphabet[char] & (1 << 11)) == (1 << 11) {
		io.WriteString(w, string(PunchcardHole))
		io.WriteString(w, MoveCursorLeft(1))
	}
	io.WriteString(w, MoveCursorDown(1))

	/* Skip frame border. */
	io.WriteString(w, MoveCursorDown(1))

	for i := 0; i < len(PunchcardDigitalLines); i++ {
		if (Alphabet[char] & (1 << i)) == (1 << i) {
			io.WriteString(w, string(PunchcardHole))
			io.WriteString(w, MoveCursorLeft(1))
		}
		io.WriteString(w, MoveCursorDown(1))
	}

	io.WriteString(w, MoveCursorRight(1))
	io.WriteString(w, MoveCursorUp(15))
}

func PrintLine(line string) {
	var buf stdbytes.Buffer

	DisplayCard()

	/* Move cursor to top left corner of the card. */
	buf.WriteString(MoveCursorUp(16))
	buf.WriteString(MoveCursorRight(1))

	for i := 0; i < len(line); i++ {
		WriteChar(&buf, line[i], i)
	}

	os.Stdout.Write(buf.Bytes())
}

func DoTab(line []byte, currPos int, destPos int) {
	for i := currPos; i < destPos; i++ {
		line[i] = ' '
		WriteChar(os.Stdout, line[i], i)
	}
}

func ClearLine() {
	/* Move to the beginning of previous line. */
	os.Stdout.Write([]byte(ESC + "[1F"))

	DisplayCard()

	/* Move cursor to top left corner of the card. */
	io.WriteString(os.Stdout, MoveCursorUp(16))
	io.WriteString(os.Stdout, MoveCursorRight(1))
}

const (
	ESC = "\033"

	Backspace = 127
)

var (
	LeftArrow  = []byte{27, 91, 68}
	RightArrow = []byte{27, 91, 67}
	Delete     = []byte{27, 91, 51, 126}
)

func WriteLines(file *os.File, lines []string) error {
	if _, err := file.Seek(0, os.SEEK_SET); err != nil {
		return fmt.Errorf("failed to rewind file: %v", err)
	}

	var length int64
	for i := 0; i < len(lines); i++ {
		if i > 0 {
			newline := []byte("\n")
			if _, err := file.Write(newline); err != nil {
				return fmt.Errorf("failed to write newline character for %dth line: %v", i, err)
			}
			length += int64(len(newline))
		}
		if _, err := io.WriteString(file, lines[i]); err != nil {
			return fmt.Errorf("failed to write %dth line: %v", i, err)
		}
		length += int64(len(lines[i]))
	}

	if err := file.Truncate(length); err != nil {
		return fmt.Errorf("failed to truncate file to length %d: %v", length, err)
	}
	return nil
}

/* TODO(anton2920): think about displaying cards stacked on each other. */
func EditFile(args []string) error {
	if len(args) != 1 {
		Usage()
	}
	file, err := os.OpenFile(args[0], os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read entire file: %v", err)
	}
	lines := stdstrings.Split(bytes.AsString(contents), "\n")

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
	if len(lines) >= 1 {
		pos = copy(line, lines[lineIndex])
	}
	PrintLine(bytes.AsString(line[:pos]))

	var quit, shouldPrint bool
	for !quit {
		buffer := make([]byte, 10)
		n, err := os.Stdin.Read(buffer)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %v", err)
		}
		buffer = buffer[:n]
		//fmt.Printf("%v\r\n", buffer)

		if len(buffer) == 1 {
			char := buffer[0]
			switch char {
			case 'q':
				os.Stdout.Write([]byte(MoveCursorDown(15)))
				os.Stdout.Write([]byte("\r\n"))
				quit = true
			case 'w':
				if err := WriteLines(file, lines); err != nil {
					return fmt.Errorf("failed to write lines to file: %v", err)
				}
			case Backspace:
				lines = strings.DeleteAt(lines, lineIndex)
				lineIndex--
				pos = len(lines[lineIndex])

				shouldPrint = true
			case '\t':
				tabPositions := SkipPositions
				for i := 0; i < len(tabPositions); i++ {
					tabPosition := tabPositions[i]
					if pos < tabPosition {
						DoTab(line, pos, tabPosition)
						pos = tabPosition
						break
					}
				}
			case '\r':
				lines[lineIndex] = string(line[:pos])
				lines = strings.InsertAt(lines, "", lineIndex+1)
				lineIndex++
				pos = 0

				shouldPrint = true
			default:
				if _, ok := Alphabet[char]; ok {
					if pos+len(buffer) < len(line) {
						if (pos == 0) && (char != 'C') {
							line[0] = ' '
							pos = 1
							os.Stdout.Write([]byte(MoveCursorRight(1)))
						}
						line[pos] = char
						WriteChar(os.Stdout, char, pos)
						pos++
					}
				}
			}
		} else if len(buffer) == 3 {
			index := lineIndex
			if stdbytes.Equal(buffer, RightArrow) {
				lineIndex++
			} else if stdbytes.Equal(buffer, LeftArrow) {
				lineIndex--
			}

			lineIndex = ints.Clamp(lineIndex, 0, len(lines))
			if index != lineIndex {
				lines[index] = string(line[:pos])
				pos = copy(line, lines[lineIndex])
				shouldPrint = true
			}
		} else if stdbytes.Equal(buffer, Delete) {
			ClearLine()
			lines[lineIndex] = ""
			pos = 0
		}

		if shouldPrint {
			fmt.Print("\r\n")
			PrintLine(lines[lineIndex])
			shouldPrint = false
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
