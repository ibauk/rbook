package main

import (
	"fmt"
	"os"

	_ "embed"

	"gopkg.in/yaml.v2"
)

const BonusSQL = `SELECT BonusID,BriefDesc,Points,IfNull(Flags,''),IfNull(Notes,''),
Cat1,Cat2,Cat3,Cat4,Cat5,Cat6,Cat7,Cat8,Cat9,IfNull(Image,''),IfNull(Waffle,''),IfNull(Coords,''),
IfNull(Question,''),IfNull(Answer,''),AskPoints
 FROM bonuses 
`

const ComboSQL = `SELECT ComboID,BriefDesc,ScoreMethod,MinimumTicks,ScorePoints,IfNull(Bonuses,''),
Cat1,Cat2,Cat3,Cat4,Cat5,Cat6,Cat7,Cat8,Cat9,Compulsory
 FROM combinations 
`

const EntrantSQL = `SELECT EntrantID,IfNull(RiderName,''),IfNull(PillionName,''),IfNull(Bike,''),IfNull(BikeReg,''),OdoKms,Cohort
 FROM entrants 
`

const htmlhead1 = `
<!DOCTYPE html>
<html lang="en">

<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1" >
<title>RBook doc</title>
<style>
`

const htmlhead2 = `
</style>
</head>

<body>
<div class="pages">
`

const htmlfoot = `
</div> <!-- pages -->
</body>
</html>
`

// const type_bonus = "bonus"
const type_combo = "combo"

const type_entrant = "entrant"

// const type_static = "static"
const stream_prefix = "stream"

//go:embed css/reboot.css
var css_reboot string

//go:embed css/a4portrait.css
var css_a4portrait string

//go:embed css/a4landscape.css
var css_a4landscape string

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
	Title               string        `yaml:"title"`
	Description         string        `yaml:"description"`
	ProjectFolder       string        `yaml:"projectFolder"`
	OutputFolder        string        `yaml:"outputFolder"`
	OutputFile          string        `yaml:"rallybookFile"`
	GPX                 GPXParams     `yaml:"generateGPX"`
	Database            string        `yaml:"database"`
	ImageFolder         string        `yaml:"imageFolder"`
	Sections            []string      `yaml:"sections"`
	Streams             []BonusStream `yaml:"streams"`
	Landscape           bool          `yaml:"landscape"`
	BonusSQL            string        `yaml:"bonusSQL"`
	ComboSQL            string        `yaml:"comboSQL"`
	EntrantSQL          string        `yaml:"entrantSQL"`
	AskPointsVarPrefix  string        `yaml:"askPointsVariablePrefix"`
	AskPointsMultPrefix string        `yaml:"askPointsMultiplierPrefix"`
}

type Bonus struct {
	BonusID                                                string
	BriefDesc                                              string
	Points                                                 string
	Flags                                                  string
	Notes                                                  string
	Waffle                                                 string
	Coords                                                 string
	Image                                                  string
	Cat1                                                   int
	Cat2                                                   int
	Cat3                                                   int
	Cat4                                                   int
	Cat5                                                   int
	Cat6                                                   int
	Cat7                                                   int
	Cat8                                                   int
	Cat9                                                   int
	NewLine                                                bool
	StreamID                                               string
	ImageFolder                                            string
	AlertT, AlertR, AlertF, AlertB, AlertD, AlertA, AlertN bool
	Question                                               string
	Answer                                                 string
	HasWaffle                                              bool
	HasNotes                                               bool
	AskPoints                                              bool
	Lat                                                    float64
	Lon                                                    float64
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

type Entrant struct {
	EntrantID   int
	RiderName   string
	PillionName string
	Bike        string
	BikeReg     string
	OdoKms      bool
	Cohort      int
	NewLine     bool
	StreamID    string
	ImageFolder string
}

func newBonus() *Bonus {

	var b Bonus

	b.Points = "1"
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

func newEntrant() *Entrant {

	var e Entrant

	return &e
}

func loadConfig() {

	configPath := *yml

	if !fileExists(configPath) {
		configPath += ".yml"
		if !fileExists(configPath) {
			fmt.Printf("Can't find config file %v\n", configPath)
			os.Exit(1)
		}
	}

	fmt.Printf("Loading config %v\n", configPath)
	file, err := os.Open(configPath)
	if err != nil {
		panic(err)
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

	if *outputGPX == "" {
		*outputGPX = CFG.GPX.OutputGPX
	}
	//fmt.Printf("CFG now reads %v\n\n", CFG.ImageFolder)
}
