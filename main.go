package main

import (
	"bufio"
	"flag"
	"fmt"
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

const PunchcardDigitalLines = `│00000│0│000000000000000000000000000000000000000000000000000000000000000000│00000000│
│11111│1│111111111111111111111111111111111111111111111111111111111111111111│11111111│
│22222│2│222222222222222222222222222222222222222222222222222222222222222222│22222222│
│33333│3│333333333333333333333333333333333333333333333333333333333333333333│33333333│
│44444│4│444444444444444444444444444444444444444444444444444444444444444444│44444444│
│55555│5│555555555555555555555555555555555555555555555555555555555555555555│55555555│
│66666│6│666666666666666666666666666666666666666666666666666666666666666666│66666666│
│77777│7│777777777777777777777777777777777777777777777777777777777777777777│77777777│
│88888│8│888888888888888888888888888888888888888888888888888888888888888888│88888888│
│99999│9│999999999999999999999999999999999999999999999999999999999999999999│99999999│`

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

func DisplayFirstLine(line string) {
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

func DisplayPunchedLine(line string, cardLine string, search int) {
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

func DisplayTwelvethLine(line string) {
	DisplayPunchedLine(line, PunchcardTwelvethLine, 1<<12)
}

func DisplayEleventhLine(line string) {
	DisplayPunchedLine(line, PunchcardEleventhLine, 1<<11)
}

func DisplayDigitalLines(line string) {
	cardLine := PunchcardDigitalLines
	nline := 0

	for {
		nl := strings.IndexByte(cardLine, '\n')
		if nl == -1 {
			DisplayPunchedLine(line, cardLine, 1<<nline)
			break
		} else {
			DisplayPunchedLine(line, cardLine[:nl], 1<<nline)
		}
		cardLine = cardLine[nl+1:]
		nline++
	}
}

func DisplayLine(line string) {
	fmt.Print(PunchcardHeader + "\r\n")
	DisplayFirstLine(line)
	DisplayTwelvethLine(line)
	fmt.Print(PunchcardHR1 + "\r\n")
	DisplayEleventhLine(line)
	fmt.Print(PunchcardHR2 + "\r\n")
	DisplayDigitalLines(line)
	fmt.Print(PunchcardFooter + "\r\n")
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
		DisplayLine(scanner.Text())
	}

	return nil
}

func EditFile(args []string) error {
	if len(args) != 1 {
		Usage()
	}
	file, err := os.OpenFile(args[0], os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

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

	_ = file

forLoop:
	for {
		DisplayLine(string(line[:pos]))

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
				break forLoop
			case 127:
				line = line[:0]
				pos = 0
			default:
				if len(buffer) <= len(line)-pos {
					pos += copy(line[pos:], buffer)
				}
			}
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
