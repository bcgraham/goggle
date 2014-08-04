package main

import (
	"math/rand"
	"fmt"
	"os"
	"log"
	"strings"
	"bufio"
	"time"
	"sort"
	"flag"
	"runtime"
	"runtime/pprof"
)

const OFFICIAL_SIZE = 4

type board [][]string

type dictionary []string 

type cell struct {
	row, col int
}

type path []cell

type words []string

type Done struct {}

type DoneSignal chan Done

var size = flag.Int("size", OFFICIAL_SIZE, "the length of one side of the boggle square")
var showBoard = flag.Bool("showBoard", false, "default false suppresses the board. true will show the board prior to solving it.")
var profile = flag.String("profile", "", "location of the profile file to be used by pprof")

func main() {
	runtime.GOMAXPROCS(4)
	flag.Parse()
	if *profile != "" {
		prof, err := os.Create(*profile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(prof)
		defer pprof.StopCPUProfile()
	}

	gridside := *size
	board := make(board, 0)
	r := rand.New(rand.NewSource(0))
	if gridside == 4 {
		board = MakeOfficialBoard(r)
	} else {
		board = MakeGenericBoard(r, gridside)
	}
	f, err := os.Open("words.txt")
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(f)
	dict := make(dictionary, 0)
	for scanner.Scan() {
		word := strings.ToLower(scanner.Text())
		if len(word)>=3 {
			dict = append(dict, word)
		}
	}
	if *showBoard {
		fmt.Println(board)
	}
	start := time.Now()

	ws := make(chan string)
	done := make(DoneSignal, len(board)*len(board))
	var workersRemaining int
	for i, _ := range board {
		for j, _ := range board {
			c := cell{i, j}
			p := path{c}
			workersRemaining++
			go Wordfinder(p, dict, board, ws, done)
		}
	}
	answer := make(words, 0)
	go func() {
		for word := range ws {
			answer = append(answer, word)
		}
	}()

	for _ = range done {
		if workersRemaining > 1 {
			workersRemaining--
		} else {
			workersRemaining--
			close(ws)
			break
		}
	}

	answer = answer.RemoveDuplicates()

	fmt.Printf("Found %v words.\n", len(answer))
	fmt.Printf("Longest %v words found were: %v\n", 10, answer.LongestNWords(10))
	fmt.Println("Solve time was: ",time.Since(start))
}

func Wordfinder(p path, d dictionary, b board, ws chan<- string, out DoneSignal) {
	if !b.IsLegal(p) {
		out <- (Done{})
		return
	}
	w := ""
	for _, c := range p {
		w += b[c.row][c.col]
	}
	subdict := d.Filter(w)
	if len(subdict)==0 {
		out <- (Done{})
		return
	}
	if subdict.Contains(w) {
		ws <- w
	}

	in := make(DoneSignal, 4)
	workersRemaining := 4
	up 		:= p.Offset(-1, 0)
	down 	:= p.Offset(1, 0)
	left 	:= p.Offset(0, -1)
	right 	:= p.Offset(0, 1)

	go Wordfinder(up, subdict, b, ws, in)
	go Wordfinder(left, subdict, b, ws, in)
	go Wordfinder(down, subdict, b, ws, in)
	go Wordfinder(right, subdict, b, ws, in)

	for _ = range in {
		if workersRemaining > 1 {
			workersRemaining--
		} else {
			workersRemaining--
			out <- (Done{})
			break
		}
	}
	return
}

func (in dictionary) Filter(prefix string) (out dictionary) {
	out = make(dictionary, 0)
	for _, w := range in {
		if len(w)>=len(prefix) {
			if w[:len(prefix)] == prefix {
				out = append(out, w)
			} else if len(out)>0 {
				return out 	
			}
		}
	}
	return out
}

func (in path) Offset(row, col int) (out path) {
	out = make(path, 0)
	out = append(out, in...)
	last := out[len(out)-1]
	out = append(out, cell{last.row+row, last.col+col})
	return out
}

func (dict dictionary) Contains(candidate string) bool {
	for _, word := range dict {
		if word==candidate {
			return true
		}
	}
	return false
}

func (board board) IsLegal(path path) bool {
	last := path[len(path)-1]
	if last.row < 0 || last.row>=len(board) {
		return false
	}
	if last.col < 0 || last.col>=len(board[0]) {
		return false
	}
	for i:=0; i<len(path)-1; i++ {
		if last == path[i] {
			return false 
		}
	}
	return true
}

func (in words) RemoveDuplicates() (out words) {
	uniques := make(map[string]Done)
	for _, w := range in {
		if _, ok := uniques[w]; !ok {
			uniques[w] = Done{}
		}
	}
	out = make(words, 0)
	for w, _ := range uniques {
		out = append(out, w)
	}
	return out
}


func MakeGenericBoard(r *rand.Rand, gridside int) (b board) {
	b = make(board, 0)
	for i:=0; len(b) < gridside; i++ {
		b = append(b, make([]string, 0))
		for j:=0; len(b[len(b)-1]) < gridside; j++ {
			b[i] = append(b[i], Letter(r))
		}
	}
	return b
}

func MakeOfficialBoard(r *rand.Rand) (b board) {
	dice := make([]string, 0)
	dice = append(dice, strings.ToLower("AAEEGN"))
	dice = append(dice, strings.ToLower("ELRTTY"))
	dice = append(dice, strings.ToLower("AOOTTW"))
	dice = append(dice, strings.ToLower("ABBJOO"))
	dice = append(dice, strings.ToLower("EHRTVW"))
	dice = append(dice, strings.ToLower("CIMOTU"))
	dice = append(dice, strings.ToLower("DISTTY"))
	dice = append(dice, strings.ToLower("EIOSST"))
	dice = append(dice, strings.ToLower("DELRVY"))
	dice = append(dice, strings.ToLower("ACHOPS"))
	dice = append(dice, strings.ToLower("HIMNQU"))
	dice = append(dice, strings.ToLower("EEINSU"))
	dice = append(dice, strings.ToLower("EEGHNW"))
	dice = append(dice, strings.ToLower("AFFKPS"))
	dice = append(dice, strings.ToLower("HLNNRZ"))
	dice = append(dice, strings.ToLower("DEILRX"))
	b = make(board, 0)
	for i:=0; len(b)<OFFICIAL_SIZE; i++ {
		b = append(b, make([]string, 0))
		for j:=0; len(b[i])<OFFICIAL_SIZE; j++ {
			numf := r.Float32()
			num := int((numf*6))
			if num < 6 {
				b[i] = append(b[i], string(dice[(i*OFFICIAL_SIZE)+j][num]))
			}
		}
	}
	return b
}

func (in words) LongestNWords(n int) (out words) {
	out = make(words, n)
	WordLoop:
	for _, w := range in {
		for i:=0; i<n; i++ {
			if len(w) > len(out[i]) {
				for j:=n-2; j>=i; j-- {
					out[j+1] = out[j]
				}
				out[i]=w
				continue WordLoop
			}
		}
	}
	return out
}

func Letter(r *rand.Rand) string {
	x := r.Float32()
	switch {
		case x <= 0.062500:
			return "a"
		case x <= 0.083333:
			return "b"
		case x <= 0.104167:
			return "c"
		case x <= 0.135417:
			return "d"
		case x <= 0.250000:
			return "e"
		case x <= 0.270833:
			return "f"
		case x <= 0.291667:
			return "g"
		case x <= 0.343750:
			return "h"
		case x <= 0.406250:
			return "i"
		case x <= 0.416667:
			return "j"
		case x <= 0.427083:
			return "k"
		case x <= 0.468750:
			return "l"
		case x <= 0.489583:
			return "m"
		case x <= 0.552083:
			return "n"
		case x <= 0.625000:
			return "o"
		case x <= 0.645833:
			return "p"
		case x <= 0.656250:
			return "q"
		case x <= 0.708333:
			return "r"
		case x <= 0.770833:
			return "s"
		case x <= 0.864583:
			return "t"
		case x <= 0.895833:
			return "u"
		case x <= 0.916667:
			return "v"
		case x <= 0.947917:
			return "w"
		case x <= 0.958333:
			return "x"
		case x <= 0.989583:
			return "y"
		case x <= 1.000000:
			return "z"
	}
	return ""
}

func (b board) String() string {
	s := ""
	if len(b) > 75 {
		s += "The board's really big, but trust me, it's here."
	} else if len(b) > 35 {
		s += strings.Repeat("-", (len(b)+2))
		s += "\n"
		for _, row := range b {
			s += "|"
			for _, cell := range row {
				s += cell
				s += ""
			}
			s += "|\n"
		}
		s += strings.Repeat("-", (len(b)+2))
	} else {
		s += strings.Repeat("-", (len(b)*2)+3)
		s += "\n"
		for _, row := range b {
			s += "| "
			for _, cell := range row {
				s += cell
				s += " "
			}
			s += "|\n"
		}
		s += strings.Repeat("-", (len(b)*2)+3)
	}
	return s
}