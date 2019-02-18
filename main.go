package main

import (
	"bytes"
	"database/sql"
	"flag"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	gpkgURLParam := flag.String("gpkgurl", "", "URL pointing to a geopackage (https://example.com/geopackage.gpkg).")
	checkParameters(gpkgURLParam)

	gpkgFile := createTmpFile()
	defer os.Remove(gpkgFile.Name())
	if errDownloadGeopackage := downloadGeopackage(gpkgFile, *gpkgURLParam); errDownloadGeopackage != nil {
		log.Fatal(errDownloadGeopackage)
	}
	log.Println("Geopackage downloaded")
	geopackage := openGeopackage(gpkgFile)
	defer geopackage.Close()
	layers := getLayersFromGeopackage(geopackage)
	for _, layer := range layers {
		columns := getPropertiesFromLayer(layer, geopackage)
		htmlBuffer := generateHTMLForLayer(layer, columns)
		writeHTMLfile(layer, htmlBuffer)
	}
}

// Check if parameters are provided
func checkParameters(gpkgURLParam *string) {
	flag.Parse()
	if *gpkgURLParam == "" {
		log.Fatal("gpkgUrl is required. Run with -h for help.")
	}
	log.Printf("Starting download for: %s", *gpkgURLParam)
}

// Create a temporary file
func createTmpFile() *os.File {
	tmpFile, errTmpFile := ioutil.TempFile(os.TempDir(), "gpkg-")
	if errTmpFile != nil {
		log.Fatal("Cannot create temporary file", errTmpFile)
	}
	log.Println("Created File: " + tmpFile.Name())
	return tmpFile
}

// Download a Geopackage and store it in a file
func downloadGeopackage(gpkgFile *os.File, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	defer gpkgFile.Close()
	_, err = io.Copy(gpkgFile, resp.Body)
	return err
}

// Open Geopackage DB
func openGeopackage(gpkgFile *os.File) *sql.DB {
	log.Println("Opening Geopackage: " + gpkgFile.Name())
	db, errDb := sql.Open("sqlite3", gpkgFile.Name())
	if errDb != nil {
		log.Fatal("Cannot open Geopackage", errDb)
	}
	return db
}

// Read layers from Geopackage
func getLayersFromGeopackage(geopackage *sql.DB) []string {
	log.Println("Searching for layers in Geopackage")
	rows, errDb := geopackage.Query("SELECT table_name from gpkg_contents")
	if errDb != nil {
		log.Fatal("Error with querying Geopackage", errDb)
	}
	var layers []string
	var layer string
	for rows.Next() {
		rows.Scan(&layer)
		log.Println("Layer found: " + layer)
		layers = append(layers, layer)
	}
	if layers == nil {
		log.Fatal("No layers found!")
	}
	return layers
}

// Read columns from layer
func getPropertiesFromLayer(layer string, geopackage *sql.DB) []string {
	log.Println("Searching for columns for layer '" + layer + "' in Geopackage")
	statement, errDb := geopackage.Query("SELECT * FROM " + layer)
	if errDb != nil {
		log.Fatal("Error with querying Geopackage", errDb)
	}
	columns, errDb := statement.Columns()
	if errDb != nil {
		log.Fatal("Error with querying Geopackage", errDb)
	}
	for _, column := range columns {
		log.Print("Column found: " + column)
	}
	return columns
}

// Generate HTML for layer
func generateHTMLForLayer(layer string, columns []string) *bytes.Buffer {
	buf := new(bytes.Buffer)
	buf.WriteString(htmlStart)
	log.Print("Generate HTML for layer: " + layer)
	layerTemplate, err := template.New("layer").Parse(htmlLayer)
	if err != nil {
		log.Fatal(err)
	}
	layerReplace := map[string]interface{}{
		"layer": template.HTML(layer),
	}
	err = layerTemplate.ExecuteTemplate(buf, "layer", layerReplace)
	if err != nil {
		log.Fatal(err)
	}
	columnHeadTemplate, err := template.New("column").Parse(htmlColumnHead)
	if err != nil {
		log.Fatal(err)
	}
	for _, column := range columns {
		if checkColumn(column) {
			columnHeadReplace := map[string]interface{}{
				"column": template.HTML(column),
			}
			err = columnHeadTemplate.ExecuteTemplate(buf, "column", columnHeadReplace)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	buf.WriteString("</tr><tr>")
	columnRowTemplate, err := template.New("column").Parse(htmlColumnRow)
	if err != nil {
		log.Fatal(err)
	}
	for _, column := range columns {
		if checkColumn(column) {
			columnRowReplace := map[string]interface{}{
				"column": template.HTML(column),
			}
			err = columnRowTemplate.ExecuteTemplate(buf, "column", columnRowReplace)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	buf.WriteString(htmlEnd)
	return buf
}

// Check if column name should be included in HTML template
func checkColumn(columnName string) bool {
	badColumns := []string{"geom", "shape_len", "shape_leng", "shape_area"}
	for _, badColumn := range badColumns {
		if strings.EqualFold(badColumn, columnName) {
			return false
		}
	}
	return true
}

// Write HTML to file
func writeHTMLfile(layer string, htmlBuffer *bytes.Buffer) {
	fileName := layer + ".html"
	errFile := ioutil.WriteFile(fileName, htmlBuffer.Bytes(), 0777)
	if errFile != nil {
		log.Fatal("Cannot create temporary file", errFile)
	}
}

const htmlStart = "<!-- MapServer FeatureInfo Template --><html><head><title>GetFeatureInfo output</title></head><style type=\"text/css\">table.featureInfo, table.featureInfo td, table.featureInfo th { border: 1px solid #ddd; border-collapse: collapse; margin: 0; padding: 0; font-size: 90%; padding: .2em .1em; } table.featureInfo th { padding: .2em .2em; font-weight: bold; background: #eee; } table.featureInfo td { background: #fff; } table.featureInfo tr.odd td { background: #eee; } table.featureInfo caption { text-align: left; font-size: 100%; font-weight: bold; padding: .2em .2em; }</style><body><table class=\"featureInfo\">"
const htmlLayer = "<caption class=\"featureInfo\">{{.layer}}</caption><tr>"
const htmlColumnHead = "<th>{{.column}}</th>"
const htmlColumnRow = "<td>[{{.column}}]</td>"
const htmlEnd = "</tr></table><br /></body></html><!-- Generated by PDOK ( https://www.pdok.nl/ ) -->"
