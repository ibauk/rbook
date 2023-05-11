package main

import (
	"fmt"
	"strings"
)

type GPXParams struct {
	OutputGPX   string `yaml:"outputFile"`
	SymbolGPX   string `yaml:"symbol"`
	LinkGPX     string `yaml:"link2map"`
	CodeOnlyGPX bool   `yaml:"bonusidOnly"`
}

const gpxheader = `<?xml version="1.0" encoding="utf-8"?>
<gpx creator="Bob Stammers (` + apptitle + `)" version="1.1"
xsi:schemaLocation="http://www.topografix.com/GPX/1/1 
http://www.topografix.com/GPX/1/1/gpx.xsd" 
xmlns="http://www.topografix.com/GPX/1/1" 
xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
`

func xmlsafe(s string) string {

	x := map[string]string{`&`: `&amp;`, `"`: `&quot;`, `<`: `&lt;`, `>`: `&gt;`, `'`: `&#39;`}
	res := s
	for k, v := range x {
		res = strings.ReplaceAll(res, k, v)
	}
	return res
}

func writeWaypoint(lat, lon float64, bonusid, briefdesc string) {

	wpt := fmt.Sprintf("<wpt lat=\"%v\" lon=\"%v\"><name>%v", lat, lon, xmlsafe(bonusid))
	if !CFG.GPX.CodeOnlyGPX {
		wpt += fmt.Sprintf("-%v", xmlsafe(briefdesc))
	}
	wpt += "</name>"
	GPXF.WriteString(wpt)
	if CFG.Title != "" {
		GPXF.WriteString(fmt.Sprintf("<cmt>%v</cmt>", xmlsafe(CFG.Title)))
	}
	if CFG.GPX.LinkGPX != "" {
		GPXF.WriteString(fmt.Sprintf(`<link href="%v%v,%v" />`, CFG.GPX.LinkGPX, lat, lon))
	}
	if CFG.GPX.SymbolGPX != "" {
		GPXF.WriteString(fmt.Sprintf("<sym>%v</sym>", CFG.GPX.SymbolGPX))
	}
	GPXF.WriteString("</wpt>\n")

}

func completeGPX() {

	GPXF.WriteString("</gpx>\n")

}
