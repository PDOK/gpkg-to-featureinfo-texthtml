package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
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
	gpkgURLParam := flag.String("gpkg-url", "", "URL pointing to a geopackage (https://example.com/geopackage.gpkg)")
	gpkgPathParam := flag.String("gpkg-path", "", "Path pointing to a geopackage (./geopackage.gpkg)")

	checkParameters(gpkgURLParam, gpkgPathParam)

	var gpkgFile *os.File
	if *gpkgURLParam != "" {
		gpkgFile = createTmpFile()
		defer os.Remove(gpkgFile.Name())
		if errDownloadGeopackage := downloadGeopackage(gpkgFile, *gpkgURLParam); errDownloadGeopackage != nil {
			log.Fatal(errDownloadGeopackage)
		}
	} else {
		var err error
		gpkgFile, err = os.Open(*gpkgPathParam)
		if err != nil {
			log.Fatal("Error opening file", gpkgPathParam)
		}
	}

	geopackage := openGeopackage(gpkgFile)
	defer geopackage.Close()
	geomColumns := getGeometryColumnsFromGeopackage(geopackage)
	layers := getLayersFromGeopackage(geopackage)
	for _, layer := range layers {
		columns := getPropertiesFromLayer(layer, geopackage)
		htmlBuffer := generateHTMLForLayer(layer, columns, geomColumns)
		writeHTMLfile(layer, htmlBuffer)
	}
}

// Check if parameters are provided
func checkParameters(gpkgURLParam *string, gpkgPathParam *string) {
	flag.Parse()
	if *gpkgURLParam == "" && *gpkgPathParam == "" {
		log.Fatal("gpkgUrl or gpkgPath is required. Run with -h for help.")
	} else if *gpkgURLParam != "" && *gpkgPathParam != "" {
		log.Fatal("either gpkgUrl or gpkgPath is required. Run with -h for help.")
	}
}

// Create a temporary file
func createTmpFile() *os.File {
	os.Mkdir("/tmp", 0777)
	tmpFile, errTmpFile := ioutil.TempFile(os.TempDir(), "gpkg-")
	if errTmpFile != nil {
		log.Fatal("Cannot create temporary file", errTmpFile)
	}
	log.Println("Created File: " + tmpFile.Name())
	return tmpFile
}

// Download a Geopackage and store it in a file
func downloadGeopackage(gpkgFile *os.File, url string) error {
	log.Printf("Starting download for: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("didn't get a 200 statuscode, instead got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	defer gpkgFile.Close()
	_, err = io.Copy(gpkgFile, resp.Body)
	log.Println("Geopackage downloaded")
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

// Read layers from Geopackage
func getGeometryColumnsFromGeopackage(geopackage *sql.DB) []string {
	log.Println("Searching for Geometry Columns in Geopackage")
	rows, errDb := geopackage.Query("SELECT DISTINCT column_name FROM gpkg_geometry_columns")
	if errDb != nil {
		log.Fatal("Error with querying Geopackage", errDb)
	}
	var columns []string
	var column string
	for rows.Next() {
		rows.Scan(&column)
		log.Println("Geometry column found: " + column)
		columns = append(columns, column)
	}
	if columns == nil {
		log.Fatal("No geometry columns found!")
	}
	return columns
}

// Generate HTML for layer
func generateHTMLForLayer(layer string, columns []string, geomColumns []string) *bytes.Buffer {
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
		if checkColumn(column, geomColumns) {
			columnHeadReplace := map[string]interface{}{
				"column": template.HTML(column),
			}
			err = columnHeadTemplate.ExecuteTemplate(buf, "column", columnHeadReplace)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	buf.WriteString("\t\t\t</tr>\n\t\t\t<tr>\n")
	columnRowTemplate, err := template.New("column").Parse(htmlColumnRow)
	if err != nil {
		log.Fatal(err)
	}
	for _, column := range columns {
		if checkColumn(column, geomColumns) {
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
func checkColumn(columnName string, geomColumns []string) bool {
	badColumns := []string{"geom", "shape_len", "shape_leng", "shape_area"}
	badColumns = append(badColumns, geomColumns...)
	for _, badColumn := range badColumns {
		if strings.EqualFold(badColumn, columnName) {
			return false
		}
	}
	return true
}

// Write HTML to file
func writeHTMLfile(layer string, htmlBuffer *bytes.Buffer) {
	os.Mkdir("output", 0777)
	fileName := "output/" + layer + ".html"
	errFile := ioutil.WriteFile(fileName, htmlBuffer.Bytes(), 0777)
	if errFile != nil {
		log.Fatal("Cannot create html file", errFile)
	}
}

const htmlStart = "<!-- MapServer Template -->\n<html>\n\t<head>\n\t\t<title>GetFeatureInfo output</title>\n\t</head>\n\t<style type=\"text/css\">table.featureInfo, table.featureInfo td, table.featureInfo th { border: 1px solid #ddd; border-collapse: collapse; margin: 0; padding: 0; font-size: 90%; padding: .2em .1em; } table.featureInfo th { padding: .2em .2em; font-weight: bold; background: #eee; } table.featureInfo td { background: #fff; } table.featureInfo tr.odd td { background: #eee; } table.featureInfo caption { text-align: left; font-size: 100%; font-weight: bold; padding: .2em .2em; }</style>\n\t<body>\n\t\t<table class=\"featureInfo\">\n"
const htmlLayer = "\t\t\t<caption class=\"featureInfo\">{{.layer}}</caption>\n\t\t\t<tr>\n"
const htmlColumnHead = "\t\t\t\t<th>{{.column}}</th>\n"
const htmlColumnRow = "\t\t\t\t<td>[{{.column}}]</td>\n"
const htmlEnd = "\t\t\t</tr>\n\t\t</table>\n\t</body>\n</html>\n<!-- Generated by PDOK ( https://www.pdok.nl/ ) -->"
