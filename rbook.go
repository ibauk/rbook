package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"net/url"
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
var outputfile = flag.String("to", "", "Output filename. Default to YAML config")

var DBH *sql.DB
var OUTF *os.File

const type_bonus = "bonus"
const type_combo = "combo"
const type_static = "static"
const stream_prefix = "stream"

type BonusStream struct {
	StreamID     string `yaml:"streamid"`
	Type         string `yaml:"type"` // bonus, combo, static
	WhereString  string `yaml:"wherestring"`
	BonusOrder   string `yaml:"bonusorder"`
	MaxPerLine   int    `yaml:"maxperline"`
	LinesPerPage int    `yaml:"linesperpage"`
	BrPerLine    int    `yaml:"brperline"`
	TemplateID   string `yaml:"template"`
	NoPageTop    bool   `yaml:"nopagetop"`
}

var CFG struct {
	Title         string        `yaml:"title"`
	Description   string        `yaml:"description"`
	ProjectFolder string        `yaml:"projectfolder"`
	OutputFolder  string        `yaml:"outputfolder"`
	OutputFile    string        `yaml:"outputfile"`
	Database      string        `yaml:"database"`
	ImageFolder   string        `yaml:"imagefolder"`
	Sections      []string      `yaml:"sections"`
	Streams       []BonusStream `yaml:"streams"`
	Landscape     bool          `yaml:"landscape"`
}

type Bonus struct {
	BonusID                                        string
	Title                                          string
	Points                                         int
	Flags                                          string
	Notes                                          string
	Waffle                                         string
	Coords                                         string
	Image                                          string
	Cat1                                           int
	Cat2                                           int
	Cat3                                           int
	Cat4                                           int
	Cat5                                           int
	Cat6                                           int
	Cat7                                           int
	Cat8                                           int
	Cat9                                           int
	NewLine                                        bool
	StreamID                                       string
	ImageFolder                                    string
	AlertT, AlertR, AlertF, AlertB, AlertD, AlertA bool
	Question                                       string
	Answer                                         string
	HasWaffle                                      bool
	HasNotes                                       bool
	AskPoints                                      bool
}

type ComboBonus struct {
	BonusID   string
	BriefDesc string
}
type Combo struct {
	ComboID      string
	BriefDesc    string
	ScoreMethod  int
	MinimumTicks int
	ScorePoints  string
	BonusList    string
	Bonuses      []ComboBonus
	Cat1         int
	Cat2         int
	Cat3         int
	Cat4         int
	Cat5         int
	Cat6         int
	Cat7         int
	Cat8         int
	Cat9         int
	Compulsory   bool
	NewLine      bool
	StreamID     string
}

func newBonus() *Bonus {

	var b Bonus

	b.Points = 1
	b.Waffle = ""
	b.Coords = ""
	b.Image = ""
	b.NewLine = false
	b.ImageFolder = CFG.ImageFolder

	return &b

}

func newCombo() *Combo {

	var b Combo

	b.Compulsory = false
	b.MinimumTicks = 0
	b.NewLine = false

	return &b

}
func fileExists(x string) bool {

	_, err := os.Stat(x)
	//return !errors.Is(err, os.ErrNotExist)
	return err == nil

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

	if *outputfile == "" {
		*outputfile = CFG.OutputFile
		if *outputfile == "" {
			fmt.Println("Must specify an outputfile name")
			os.Exit(1)
		}
	}
	fmt.Printf("CFG now reads %v\n\n", CFG.ImageFolder)
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
					//fmt.Printf("Calling combos %v\n", v.StreamID)
					emitCombos(sx, sf[1])
				} else {
					//fmt.Printf("Calling bonuses %v\n", v.StreamID)
					emitBonuses(sx, sf[1], v.NoPageTop)
				}
			}
		}

	}

}

func emitBonuses(s int, sf string, nopage bool) {

	sql := "SELECT BonusID,BriefDesc,Points,IfNull(Flags,''),IfNull(Notes,''),"
	sql += "Cat1,Cat2,Cat3,Cat4,Cat5,Cat6,Cat7,Cat8,Cat9,IfNull(Image,''),IfNull(Waffle,''),IfNull(Coords,''),"
	sql += "IfNull(Question,''),IfNull(Answer,''),AskPoints"
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
	NLines := -1
	if nopage {
		OUTF.WriteString("\n<div class='nopage'> <!-- no page -->\n")
	} else {
		OUTF.WriteString("\n<div class='page'>\n")
	}
	for rows.Next() {
		B := newBonus()
		askPoints := 0

		err := rows.Scan(&B.BonusID, &B.Title, &B.Points, &B.Flags, &B.Notes,
			&B.Cat1, &B.Cat2, &B.Cat3, &B.Cat4, &B.Cat5, &B.Cat6, &B.Cat7, &B.Cat8, &B.Cat9, &B.Image, &B.Waffle, &B.Coords,
			&B.Question, &B.Answer, &askPoints)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		B.StreamID = CFG.Streams[s].StreamID
		B.HasWaffle = B.Waffle != ""
		B.HasNotes = B.Notes != ""
		B.AskPoints = askPoints != 0

		u := url.QueryEscape(B.Image)
		//fmt.Printf("Parsed %v; got %v\n", B.Image, u)
		B.Image = u

		B.NewLine = NRex%CFG.Streams[s].MaxPerLine == 0
		if B.NewLine {
			NLines++
			if NLines >= CFG.Streams[s].LinesPerPage {
				xx := fmt.Sprintf("Nrex=%v MPL=%v NL=%v NLines=%v LPP=%v", NRex, CFG.Streams[s].MaxPerLine,
					B.NewLine, NLines, CFG.Streams[s].LinesPerPage)
				OUTF.WriteString("</div><!-- autopage -->\n<div class='page'><!-- " + xx + " -->\n")
				NLines = 0
			}
		}

		NRex++

		B.ImageFolder = CFG.ImageFolder

		setFlags(B)

		streamTemplate := sf
		if CFG.Streams[s].TemplateID != "" {
			streamTemplate = CFG.Streams[s].TemplateID
		}
		xfile := filepath.Join(CFG.ProjectFolder, streamTemplate+".html")
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
	}
	if NLines < CFG.Streams[s].LinesPerPage {
		NLines++
		n := (CFG.Streams[s].LinesPerPage - NLines) * CFG.Streams[s].BrPerLine
		OUTF.WriteString("\n<!-- " + fmt.Sprintf("NL=%v, LPP=%v, n=%v", NLines, CFG.Streams[s].LinesPerPage, n) + " -->\n")
		OUTF.WriteString("<p>" + strings.Repeat("<br>", n) + "</p>")

	}
	OUTF.WriteString("</div>")
	fmt.Printf("%v bonus records processed\n", NRex)
	rows.Close()

}

func emitCombos(s int, sf string) {

	sql := "SELECT ComboID,BriefDesc,ScoreMethod,MinimumTicks,ScorePoints,IfNull(Bonuses,''),"
	sql += "Cat1,Cat2,Cat3,Cat4,Cat5,Cat6,Cat7,Cat8,Cat9,Compulsory"
	sql += " FROM combinations "
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
	NLines := -1
	OUTF.WriteString("<div class='page'>")
	for rows.Next() {

		B := newCombo()

		err := rows.Scan(&B.ComboID, &B.BriefDesc, &B.ScoreMethod, &B.MinimumTicks, &B.ScorePoints, &B.BonusList,
			&B.Cat1, &B.Cat2, &B.Cat3, &B.Cat4, &B.Cat5, &B.Cat6, &B.Cat7, &B.Cat8, &B.Cat9, &B.Compulsory)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		B.StreamID = CFG.Streams[s].StreamID

		B.NewLine = NRex%CFG.Streams[s].MaxPerLine == 0
		if B.NewLine {
			NLines++
			if NLines >= CFG.Streams[s].LinesPerPage {
				xx := fmt.Sprintf("Nrex=%v MPL=%v NL=%v NLines=%v LPP=%v", NRex, CFG.Streams[s].MaxPerLine,
					B.NewLine, NLines, CFG.Streams[s].LinesPerPage)
				OUTF.WriteString("</div><!-- autopage -->\n<div class='page'><!-- " + xx + " -->\n")
				NLines = 0
			}
		}

		NRex++

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
	}
	if NLines < CFG.Streams[s].LinesPerPage {
		NLines++
		n := (CFG.Streams[s].LinesPerPage - NLines) * CFG.Streams[s].BrPerLine
		OUTF.WriteString("\n<!-- " + fmt.Sprintf("NL=%v, LPP=%v, n=%v", NLines, CFG.Streams[s].LinesPerPage, n) + " -->\n")
		OUTF.WriteString("<p>" + strings.Repeat("<br>", n) + "</p>")
	}

	OUTF.WriteString("</div>")
	fmt.Printf("%v Combo records processed\n", NRex)
	rows.Close()

}

func emitTopTail(F *os.File, xfile string) {

	if !fileExists(xfile) {
		return
	}
	html, err := template.ParseFiles(xfile)
	if err != nil {
		fmt.Printf("Error (%v) in static file %v\n", err, xfile)
	}
	err = html.Execute(F, CFG)
	if err != nil {
		fmt.Printf("x %v\n", err)
	}

}

func setFlags(b *Bonus) {

	for _, c := range b.Flags {
		switch c {
		case 'F':
			b.AlertF = true
		case 'T':
			b.AlertT = true
		case 'B':
			b.AlertB = true
		case 'A':
			b.AlertA = true
		case 'R':
			b.AlertR = true
		case 'D':
			b.AlertD = true
		}
	}
}
