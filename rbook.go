package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	yaml "gopkg.in/yaml.v2"
)

const apptitle = "RBook v1.0"
const progdesc = `
I print rally books using data supplied by Rallymasters in a standard format
`

var yml = flag.String("cfg", "rbook.yml", "Name of the YAML configuration")
var showusage = flag.Bool("?", false, "Show this help")
var outputfile = flag.String("to", "output.html", "Output filename")

var DBH *sql.DB
var OUTF *os.File

const type_bonus = "bonus"
const type_combo = "combo"
const type_static = "static"
const stream_prefix = "stream"

type BonusStream struct {
	StreamID    string `yaml:"streamid"`
	Type        string `yaml:"type"` // bonus, combo, static
	WhereString string `yaml:"wherestring"`
	BonusOrder  string `yaml:"bonusorder"`
	MaxPerLine  int    `yaml:"maxperline"`
}

var CFG struct {
	Title         string        `yaml:"title"`
	Description   string        `yaml:"description"`
	ProjectFolder string        `yaml:"projectfolder"`
	OutputFolder  string        `yaml:"outputfolder"`
	Database      string        `yaml:"database"`
	Imagefolder   string        `yaml:"imagefolder"`
	Sections      []string      `yaml:"sections"`
	Streams       []BonusStream `yaml:"streams"`
	Landscape     bool          `yaml:"landscape"`
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
	ImagePath  string
}

func newBonus() *Bonus {

	var b Bonus

	b.Points = 1
	b.Waffle = ""
	b.Coords = ""
	b.Image = ""
	b.ClearFloat = false
	b.ImagePath = CFG.Imagefolder

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

	configPath := *yml

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

	fmt.Printf("CFG now reads %v\n\n", CFG.Imagefolder)
}
func main() {

	var xfile string

	fmt.Printf("%v\nCopyright (c) 2022 Bob Stammers\n", apptitle)
	xfile = filepath.Join(CFG.OutputFolder, *outputfile)
	fmt.Printf("\nBook title: %v\n%v\nGenerating %v \n", CFG.Title, CFG.Description, xfile)
	OUTF, _ = os.Create(filepath.Join(CFG.OutputFolder, *outputfile))
	defer OUTF.Close()
	var err error
	DBH, err = sql.Open("sqlite3", CFG.Database)
	if err != nil {
		panic(err)
	}
	defer DBH.Close()

	//fmt.Printf("YAML:\n%v\n\n", CFG)

	for i := 0; i < len(CFG.Sections); i++ {

		sf := strings.Split(CFG.Sections[i], ".")
		if len(sf) < 2 || sf[0] != stream_prefix {

			xfile = filepath.Join(CFG.ProjectFolder, sf[0]+".html")

			emitTopTail(OUTF, xfile)
			continue
		}
		for sx, v := range CFG.Streams {
			if v.StreamID == sf[1] {
				if v.Type == type_combo {
					emitCombos(sx, sf[1])
				} else {
					emitBonuses(sx, sf[1])
				}
				break
			}
		}

	}

}

func emitBonuses(s int, sf string) {

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

			B.ClearFloat = NRex%CFG.Streams[s].MaxPerLine == 0
			B.ImagePath = CFG.Imagefolder

			xfile := filepath.Join(CFG.ProjectFolder, sf+".html")
			if !fileExists(xfile) {
				fmt.Printf("Stream %v has no template %v\n", CFG.Streams[s].StreamID, xfile)
				rows.Close()
				return
			}
			t, err := template.ParseFiles(xfile)
			if err != nil {
				fmt.Printf("Parsing error (%v) in %v\n", err, xfile)
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
}

func emitCombos(s int, sf string) {

}

func emitTopTail(F *os.File, xfile string) {

	html, err := template.ParseFiles(xfile)
	if err != nil {
		fmt.Printf("Error (%v) in static file %v\n", err, xfile)
	}
	err = html.Execute(F, CFG)
	if err != nil {
		fmt.Printf("x %v\n", err)
	}

}
