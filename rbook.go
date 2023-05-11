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

	"github.com/flopp/go-coordsparser"
	_ "github.com/mattn/go-sqlite3"
)

const apptitle = "RBook v1.3"
const progdesc = `
I print rally books using data supplied by Rallymasters in a standard format
`

var yml = flag.String("cfg", "rbook.yml", "Name of the YAML configuration")
var showusage = flag.Bool("?", false, "Show this help")
var outputfile = flag.String("book", "", "Output filename. Default to YAML config")
var outputGPX = flag.String("gpx", "", "Output GPX. Default to YAML config")

var DBH *sql.DB
var OUTF *os.File
var GPXF *os.File

// const type_bonus = "bonus"
const type_combo = "combo"

const type_entrant = "entrant"

// const type_static = "static"
const stream_prefix = "stream"

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

func main() {

	var xfile string

	fmt.Printf("%v\nCopyright (c) 2023 Bob Stammers\n", apptitle)
	if *outputfile != "" && *outputfile != "none" {
		xfile = filepath.Join(CFG.OutputFolder, *outputfile)
		fmt.Printf("\nBook title: %v\n%v\nGenerating %v \n", CFG.Title, CFG.Description, xfile)
		OUTF, _ = os.Create(xfile)
		defer OUTF.Close()
	}

	if *outputGPX != "" {
		yfile := filepath.Join(CFG.OutputFolder, *outputGPX)
		fmt.Printf("Generating GPX %v\n", yfile)
		GPXF, _ = os.Create(yfile)
		defer GPXF.Close()
		GPXF.WriteString(gpxheader)
	}

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

			if OUTF != nil {
				emitTopTail(OUTF, xfile)
			}
			continue
		}
		for sx, v := range CFG.Streams {
			if v.StreamID == sf[1] {
				if v.Type == type_combo {
					//fmt.Printf("Calling combos %v\n", v.StreamID)
					emitCombos(sx, sf[1])
				} else if v.Type == type_entrant {
					emitEntrants(sx, sf[1], v.NoPageTop)
				} else {
					//fmt.Printf("Calling bonuses %v\n", v.StreamID)
					emitBonuses(sx, sf[1], v.NoPageTop)
				}
			}
		}

	}
	if GPXF != nil {
		completeGPX()
	}

}

func emitBonuses(s int, sf string, nopage bool) {

	var sql string
	if CFG.BonusSQL != "" {
		sql = CFG.BonusSQL
	} else {
		sql = BonusSQL
	}
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
	if OUTF != nil {
		if nopage {
			OUTF.WriteString("\n<div class='nopage'> <!-- no page -->\n")
		} else {
			OUTF.WriteString("\n<div class='page'>\n")
		}
	}
	for rows.Next() {
		B := newBonus()
		askPoints := 0

		err := rows.Scan(&B.BonusID, &B.BriefDesc, &B.Points, &B.Flags, &B.Notes,
			&B.Cat1, &B.Cat2, &B.Cat3, &B.Cat4, &B.Cat5, &B.Cat6, &B.Cat7, &B.Cat8, &B.Cat9, &B.Image, &B.Waffle, &B.Coords,
			&B.Question, &B.Answer, &askPoints)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		B.StreamID = CFG.Streams[s].StreamID
		B.HasWaffle = B.Waffle != ""
		B.HasNotes = B.Notes != ""
		B.AskPoints = askPoints != 0
		if GPXF != nil {
			B.Lat, B.Lon, err = coordsparser.Parse(strings.ReplaceAll(strings.ReplaceAll(B.Coords, "°", " "), "'", " "))
			if err != nil {
				fmt.Printf("%v Coords err:%v\n", B.BonusID, err)
			} else {
				writeWaypoint(B.Lat, B.Lon, B.BonusID, B.BriefDesc)
			}
		}

		u := url.QueryEscape(B.Image)
		//fmt.Printf("Parsed %v; got %v\n", B.Image, u)
		B.Image = u

		if CFG.Streams[s].MaxPerLine > 0 {
			B.NewLine = NRex%CFG.Streams[s].MaxPerLine == 0
			if B.NewLine {
				NLines++
				if NLines >= CFG.Streams[s].LinesPerPage {
					xx := fmt.Sprintf("Nrex=%v MPL=%v NL=%v NLines=%v LPP=%v", NRex, CFG.Streams[s].MaxPerLine,
						B.NewLine, NLines, CFG.Streams[s].LinesPerPage)
					if OUTF != nil {
						OUTF.WriteString("</div><!-- autopage -->\n<div class='page'><!-- " + xx + " -->\n")
					}
					NLines = 0
				}
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
		if OUTF != nil {
			err = t.Execute(OUTF, B)
			if err != nil {
				fmt.Printf("x %v\n", err)
			}
		}
	}
	if NLines < CFG.Streams[s].LinesPerPage {
		NLines++
		n := (CFG.Streams[s].LinesPerPage - NLines) * CFG.Streams[s].BrPerLine
		if OUTF != nil {
			OUTF.WriteString("\n<!-- " + fmt.Sprintf("NL=%v, LPP=%v, n=%v", NLines, CFG.Streams[s].LinesPerPage, n) + " -->\n")
			OUTF.WriteString("<p>" + strings.Repeat("<br>", n) + "</p>")
		}

	}
	if OUTF != nil {
		OUTF.WriteString("</div>")
	}
	fmt.Printf("%v bonus records processed\n", NRex)
	rows.Close()

}

func emitCombos(s int, sf string) {

	var sql string
	if CFG.ComboSQL != "" {
		sql = CFG.ComboSQL
	} else {
		sql = ComboSQL
	}
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
	if OUTF != nil {
		OUTF.WriteString("<div class='page'>")
	}
	for rows.Next() {

		B := newCombo()

		err := rows.Scan(&B.ComboID, &B.BriefDesc, &B.ScoreMethod, &B.MinimumTicks, &B.ScorePoints, &B.BonusList,
			&B.Cat1, &B.Cat2, &B.Cat3, &B.Cat4, &B.Cat5, &B.Cat6, &B.Cat7, &B.Cat8, &B.Cat9, &B.Compulsory)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		B.StreamID = CFG.Streams[s].StreamID

		if CFG.Streams[s].MaxPerLine > 0 {
			B.NewLine = NRex%CFG.Streams[s].MaxPerLine == 0
			if B.NewLine {
				NLines++
				if NLines >= CFG.Streams[s].LinesPerPage {
					xx := fmt.Sprintf("Nrex=%v MPL=%v NL=%v NLines=%v LPP=%v", NRex, CFG.Streams[s].MaxPerLine,
						B.NewLine, NLines, CFG.Streams[s].LinesPerPage)
					if OUTF != nil {
						OUTF.WriteString("</div><!-- autopage -->\n<div class='page'><!-- " + xx + " -->\n")
					}
					NLines = 0
				}
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
		if OUTF != nil {
			err = t.Execute(OUTF, B)
			if err != nil {
				fmt.Printf("x %v\n", err)
			}
		}
	}
	if NLines < CFG.Streams[s].LinesPerPage {
		NLines++
		n := (CFG.Streams[s].LinesPerPage - NLines) * CFG.Streams[s].BrPerLine
		if OUTF != nil {
			OUTF.WriteString("\n<!-- " + fmt.Sprintf("NL=%v, LPP=%v, n=%v", NLines, CFG.Streams[s].LinesPerPage, n) + " -->\n")
			OUTF.WriteString("<p>" + strings.Repeat("<br>", n) + "</p>")
		}
	}
	if OUTF != nil {
		OUTF.WriteString("</div>")
	}
	fmt.Printf("%v Combo records processed\n", NRex)
	rows.Close()

}

func emitEntrants(s int, sf string, nopage bool) {

	var sql string
	if CFG.EntrantSQL != "" {
		sql = CFG.EntrantSQL
	} else {
		sql = EntrantSQL
	}
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
	if OUTF != nil {
		if nopage {
			OUTF.WriteString("\n<div class='nopage'> <!-- no page -->\n")
		} else {
			OUTF.WriteString("\n<div class='page'>\n")
		}
	}
	for rows.Next() {
		E := newEntrant()
		odoKms := 0

		err := rows.Scan(&E.EntrantID, &E.RiderName, &E.PillionName, &E.Bike, &E.BikeReg, &odoKms, &E.Cohort)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		E.StreamID = CFG.Streams[s].StreamID
		E.OdoKms = odoKms != 0

		if CFG.Streams[s].MaxPerLine > 0 {
			E.NewLine = NRex%CFG.Streams[s].MaxPerLine == 0
			if E.NewLine {
				NLines++
				if NLines >= CFG.Streams[s].LinesPerPage {
					xx := fmt.Sprintf("Nrex=%v MPL=%v NL=%v NLines=%v LPP=%v", NRex, CFG.Streams[s].MaxPerLine,
						E.NewLine, NLines, CFG.Streams[s].LinesPerPage)
					if OUTF != nil {
						OUTF.WriteString("</div><!-- autopage -->\n<div class='page'><!-- " + xx + " -->\n")
					}
					NLines = 0
				}
			}
		}

		NRex++

		E.ImageFolder = CFG.ImageFolder

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
		if OUTF != nil {
			err = t.Execute(OUTF, E)
			if err != nil {
				fmt.Printf("x %v\n", err)
			}
		}
	}
	if NLines < CFG.Streams[s].LinesPerPage {
		NLines++
		n := (CFG.Streams[s].LinesPerPage - NLines) * CFG.Streams[s].BrPerLine
		if OUTF != nil {
			OUTF.WriteString("\n<!-- " + fmt.Sprintf("NL=%v, LPP=%v, n=%v", NLines, CFG.Streams[s].LinesPerPage, n) + " -->\n")
			OUTF.WriteString("<p>" + strings.Repeat("<br>", n) + "</p>")
		}

	}
	if OUTF != nil {
		OUTF.WriteString("</div>")
	}
	fmt.Printf("%v entrant records processed\n", NRex)
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
		case 'N':
			b.AlertN = true
		}
	}
}
