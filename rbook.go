package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	yaml "gopkg.in/yaml.v2"
)

const apptitle = "RBook v1.0"
const progdesc = `
I print rally books using data supplied by Rallymasters in a standard format
`

var a5 = flag.Bool("a5", false, "Format for A5 rather than A4")
var yml = flag.String("cfg", "rbook.yml", "Name of the YAML configuration")
var doc = flag.String("project", "sample", "The name of the project to be produced")
var showusage = flag.Bool("?", false, "Show this help")
var outputfile = flag.String("to", "output.html", "Output filename")

var DBH *sql.DB
var OUTF *os.File

var projectFolder string = "projects"

type BonusStream struct {
	StreamID    string `yaml:"streamid"`
	WhereString string `yaml:"wherestring"`
	BonusOrder  string `yaml:"bonusorder"`
	MaxPerLine  int    `yaml:"maxperline"`
}

var CFG struct {
	Event       string        `yaml:"event"`
	Database    string        `yaml:"database"`
	Imagefolder string        `yaml:"imagefolder"`
	Headers     []string      `yaml:"headers"`
	Streams     []BonusStream `yaml:"streams"`
}

type Bonus struct {
	BonusID    string
	Title      string
	Points     int
	Flags      string
	Notes      string
	Waffle     string
	Coords     string
	Image      string
	Cat1       int
	Cat2       int
	Cat3       int
	Cat4       int
	Cat5       int
	Cat6       int
	Cat7       int
	Cat8       int
	Cat9       int
	ClearFloat bool
	StreamID   string
}

func newBonus() *Bonus {

	var b Bonus

	b.Points = 1
	b.Waffle = ""
	b.Coords = ""
	b.Image = ""
	b.ClearFloat = false

	return &b

}

func fileExists(x string) bool {

	_, err := os.Stat(x)
	return !errors.Is(err, os.ErrNotExist)

}

func init() {

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%v\n", apptitle)
		fmt.Fprintf(w, "%v\n", progdesc)
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showusage {
		flag.Usage()
		os.Exit(1)
	}
	loadConfig()
}

func loadConfig() {

	configPath := *yml + ".yml"

	if !fileExists(configPath) {
		fmt.Printf("Can't find config file %v\n", configPath)
		return
	}

	file, err := os.Open(configPath)
	if err != nil {
		return
	}
	defer file.Close()

	D := yaml.NewDecoder(file)
	err = D.Decode(&CFG)
	if err != nil {
		panic(err)
	}

}
func main() {

	fmt.Printf("%v\nCopyright (c) 2022 Bob Stammers\n", apptitle)
	fmt.Printf("Event: %v\nGenerating %v into %v\n", CFG.Event, *doc, *outputfile)
	OUTF, _ = os.Create(*outputfile)
	defer OUTF.Close()
	var err error
	DBH, err = sql.Open("sqlite3", CFG.Database)
	if err != nil {
		panic(err)
	}
	defer DBH.Close()

	//fmt.Printf("YAML:\n%v\n\n", CFG)
	var xfile string

	for i := 0; i < len(CFG.Headers); i++ {

		xfile = filepath.Join(projectFolder, *doc, CFG.Headers[i]+".html")
		emitTopTail(OUTF, xfile)

	}

	for s := 0; s < len(CFG.Streams); s++ {
		sql := "SELECT BonusID,BriefDesc,Points,IfNull(Flags,''),IfNull(Notes,''),"
		sql += "Cat1,Cat2,Cat3,Cat4,Cat5,Cat6,Cat7,Cat8,Cat9"
		sql += " FROM bonuses "
		if CFG.Streams[s].WhereString != "" {
			sql += " WHERE " + CFG.Streams[s].WhereString
		}
		if CFG.Streams[s].BonusOrder != "" {
			sql += " ORDER BY " + CFG.Streams[s].BonusOrder
		}
		//fmt.Printf("%v\n", sql)
		rows, err := DBH.Query(sql)
		if err != nil {
			fmt.Printf("ERROR! %v\nproduced %v\n", sql, err)
			return
		}
		NRex := 0
		for rows.Next() {
			B := newBonus()

			err := rows.Scan(&B.BonusID, &B.Title, &B.Points, &B.Flags, &B.Notes,
				&B.Cat1, &B.Cat2, &B.Cat3, &B.Cat4, &B.Cat5, &B.Cat6, &B.Cat7, &B.Cat8, &B.Cat9)
			if err != nil {
				fmt.Printf("%v\n", err)
			}

			B.StreamID = CFG.Streams[s].StreamID

			if *a5 {
				B.ClearFloat = NRex%2 == 0
			}
			xfile = filepath.Join(projectFolder, *doc, CFG.Streams[s].StreamID+".html")
			if !fileExists(xfile) {
				fmt.Printf("Stream %v has not template %v\n", CFG.Streams[s].StreamID, xfile)
				rows.Close()
				continue
			}
			t, err := template.ParseFiles(xfile)
			if err != nil {
				fmt.Printf("new %v\n", err)
			}
			err = t.Execute(OUTF, B)
			if err != nil {
				fmt.Printf("x %v\n", err)
			}
			NRex++
		}
		fmt.Printf("%v bonus records processed\n", NRex)
		rows.Close()
	}
	xfile = filepath.Join(projectFolder, *doc, "footer.html")
	emitTopTail(OUTF, xfile)
}

func emitTopTail(F *os.File, xfile string) {

	html, err := os.ReadFile(xfile)
	if err != nil {
		fmt.Printf("new %v\n", err)
	}
	F.WriteString(string(html))
}
