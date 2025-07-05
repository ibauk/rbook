package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/flopp/go-coordsparser"
	_ "github.com/mattn/go-sqlite3"
)

const apptitle = "RBook v1.8"
const progdesc = `
I print rally books using data supplied by Rallymasters in a standard format
`

// ScoreMaster AskPoints values
const smAskPointsVar = 1
const smAskPointsMult = 2

var yml = flag.String("cfg", "std.yml", "Name of the YAML configuration")
var showusage = flag.Bool("?", false, "Show this help")
var outputfile = flag.String("book", "", "Output filename. Default to YAML config")
var outputGPX = flag.String("gpx", "", "Output GPX. Default to YAML config")
var database = flag.String("db", "", "ScoreMaster database")
var verbose = flag.Bool("v", false, "verbose mode")

var DBH *sql.DB
var OUTF *os.File
var GPXF *os.File

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}

func fileExists(x string) bool {

	_, err := os.Stat(x)
	//return !errors.Is(err, os.ErrNotExist)
	return err == nil

}

func getStringFromDB(sqlx string, defval string) string {

	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	if rows.Next() {
		var res string
		err = rows.Scan(&res)
		checkerr(err)
		return res
	}
	return defval
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

	fmt.Printf("%v   Copyright (c) 2025 Bob Stammers\n", apptitle)

	fmt.Printf("Project folder is %v\n", CFG.ProjectFolder)

	if *database != "" {
		CFG.Database = *database
	}

	var err error
	DBH, err = sql.Open("sqlite3", CFG.Database)
	checkerr(err)
	defer DBH.Close()

	CFG.Title = getStringFromDB("SELECT RallyTitle FROM rallyparams", CFG.Title)

	if *outputfile != "" && *outputfile != "none" {
		if strings.ContainsRune(*outputfile, filepath.Separator) {
			xfile = *outputfile
		} else {
			xfile = filepath.Join(CFG.OutputFolder, *outputfile)
		}
		fmt.Printf("\nBook title: %v\n%v\nGenerating %v \n", CFG.Title, CFG.Description, xfile)
		OUTF, _ = os.Create(xfile)
		defer OUTF.Close()
	}

	if *outputGPX == "" {
		*outputGPX = CFG.GPX.OutputGPX
	}

	if *outputGPX != "" {
		yfile := ""
		if strings.ContainsRune(*outputGPX, filepath.Separator) {
			yfile = *outputGPX
		} else {
			yfile = filepath.Join(CFG.OutputFolder, *outputGPX)
		}

		fmt.Printf("Generating GPX %v\n", yfile)
		fmt.Print("Bonuses in config streams including 'emitGPX: true' are emitted to GPX\n")
		GPXF, _ = os.Create(yfile)
		defer GPXF.Close()
		GPXF.WriteString(gpxheader)
	}

	fmt.Println()

	fmt.Fprint(OUTF, strings.ReplaceAll(htmlhead1, "RBook doc", CFG.Title))
	fmt.Fprint(OUTF, css_reboot)
	if CFG.Landscape {
		fmt.Fprint(OUTF, css_a4landscape)
	} else {
		fmt.Fprint(OUTF, css_a4portrait)
	}
	emitTopTail(OUTF, filepath.Join(CFG.ProjectFolder, "document.css"))
	fmt.Fprint(OUTF, htmlhead2)

	for i := 0; i < len(CFG.Sections); i++ {

		sf := strings.Split(CFG.Sections[i], ".")
		if len(sf) < 2 || sf[0] != stream_prefix {

			xfile = filepath.Join(CFG.ProjectFolder, sf[0]+".html")

			if OUTF != nil {
				if *verbose {
					fmt.Printf("Emitting %v\n", sf[0])
				}
				emitTopTail(OUTF, xfile)
			}
			continue
		}
		for sx, v := range CFG.Streams {
			if v.StreamID == sf[1] {
				if *verbose {
					fmt.Printf("Streaming %v\n", sf[1])
				}
				fmt.Fprint(OUTF, `<div class="stream`+sf[1]+`">`)
				switch v.Type {
				case type_combo:
					emitCombos(sx, sf[1])
				case type_entrant:
					emitEntrants(sx, sf[1], v.NoPageTop)
				default:
					emitBonuses(sx, sf[1], v.NoPageTop, v.EmitGPX)
				}
				fmt.Fprint(OUTF, `</div>`)
			}
		}

	}
	fmt.Fprint(OUTF, htmlfoot)
	if GPXF != nil {
		completeGPX()
	}

}

func emitBonuses(s int, sf string, nopage bool, emitGPX bool) {

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
	NGpx := 0
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
		PointsVal := 0

		err := rows.Scan(&B.BonusID, &B.BriefDesc, &PointsVal, &B.Flags, &B.Notes,
			&B.Cat1, &B.Cat2, &B.Cat3, &B.Cat4, &B.Cat5, &B.Cat6, &B.Cat7, &B.Cat8, &B.Cat9, &B.Image, &B.Waffle, &B.Coords,
			&B.Question, &B.Answer, &askPoints)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		B.StreamID = CFG.Streams[s].StreamID
		B.HasWaffle = B.Waffle != ""
		B.HasNotes = B.Notes != ""
		B.AskPoints = askPoints == smAskPointsVar
		switch askPoints {
		case smAskPointsVar:
			B.Points = CFG.AskPointsVarPrefix + strconv.Itoa(PointsVal)
		case smAskPointsMult:
			B.Points = CFG.AskPointsMultPrefix + strconv.Itoa(PointsVal)
		default:
			B.Points = strconv.Itoa(PointsVal)
		}
		if GPXF != nil && emitGPX {
			B.Lat, B.Lon, err = coordsparser.Parse(cleanCoords(B.Coords))
			if err != nil {
				fmt.Printf("%v Coords err:%v\n", B.BonusID, err)
			} else {
				writeWaypoint(B.Lat, B.Lon, B.BonusID, B.BriefDesc, PointsVal)
				NGpx++
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
	}
	OUTF.WriteString("</div>")
	fmt.Printf("\n%v bonus records processed [%v]\n", NRex, sf)
	if *outputGPX != "" {
		fmt.Printf("%v bonuses included in GPX\n", NGpx)
	}
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
	if OUTF != nil && false {
		OUTF.WriteString("<div class='page'>")
	}
	for rows.Next() {

		B := newCombo()

		err := rows.Scan(&B.ComboID, &B.BriefDesc, &B.ScoreMethod, &B.MinimumTicks, &B.ScorePoints, &B.BonusList,
			&B.Cat1, &B.Cat2, &B.Cat3, &B.Cat4, &B.Cat5, &B.Cat6, &B.Cat7, &B.Cat8, &B.Cat9, &B.Compulsory)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		if B.MinimumTicks > 0 {
			expandComboPoints(B)
			//fmt.Printf("%v %v %v\n", B.ComboID, B.MinimumTicks, B.ScorePoints)
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
	}
	if OUTF != nil {
		OUTF.WriteString("</div>")
	}
	fmt.Printf("%v Combo records processed [%v]\n", NRex, sf)
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
		fmt.Printf("emitTopTail [%v] %v\n", xfile, err)
	}

}

func expandComboPoints(B *Combo) {

	// This expects to have a properly completed BonusList and ScorePoints
	if B.MinimumTicks < 1 {
		return
	}
	bl := strings.Split(B.BonusList, ",")
	sp := strings.Split(B.ScorePoints, ",")
	x := ""
	for n := B.MinimumTicks; n <= len(bl); n++ {
		//fmt.Printf("%v now\n", n)
		if x != "" {
			x += ","
		}
		x += fmt.Sprintf("%v=%v", n, sp[n-B.MinimumTicks])
	}
	B.ScorePoints = x
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
