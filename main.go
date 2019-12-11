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
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	startTime := time.Now()
	gpkgURLParam := flag.String("gpkg-url", "", "URL pointing to a geopackage (https://example.com/geopackage.gpkg)")
	gpkgPathParam := flag.String("gpkg-path", "", "Path pointing to a geopackage (./geopackage.gpkg)")
	checkParameters(gpkgURLParam, gpkgPathParam)
	gpkgFile := getGpkgFile(gpkgURLParam, gpkgPathParam)
	geopackage := openGeopackage(gpkgFile)
	defer gpkgFile.Close()
	defer geopackage.Close()
	geomColumns := getGeometryColumnsFromGeopackage(geopackage)
	layers := getLayersFromGeopackage(geopackage)
	for _, layer := range layers {
		columns := getPropertiesFromLayer(layer, geopackage)
		htmlBuffer := generateHTMLForLayer(layer, columns, geomColumns)
		writeHTMLfile(layer, htmlBuffer)
	}
	cleanup(gpkgFile, gpkgURLParam)
	programFinishedSuccesfully(startTime)
}

// Check if parameters are provided
func checkParameters(gpkgURLParam *string, gpkgPathParam *string) {
	flag.Parse()
	if *gpkgURLParam == "" && *gpkgPathParam == "" {
		log.Fatal("Error: gpkg-url or gpkg-path is required. Run with -h for help.")
	} else if *gpkgURLParam != "" && *gpkgPathParam != "" {
		log.Fatal("Error: either gpkg-url or gpkg-path is required. Run with -h for help.")
	}
}

// Create a temporary file
func createTmpFile() *os.File {
	if _, err := os.Stat("/tmp"); os.IsNotExist(err) {
		os.Mkdir("/tmp", 0777)
	}
	tmpFile, errTmpFile := ioutil.TempFile(os.TempDir(), "gpkg-")
	if errTmpFile != nil {
		log.Fatal("Cannot create temporary file", errTmpFile)
	}
	log.Println("Created temporary file: " + tmpFile.Name())
	return tmpFile
}

// Get the Geopackage as a file
func getGpkgFile(gpkgURLParam *string, gpkgPathParam *string) *os.File {
	var gpkgFile *os.File
	if *gpkgURLParam != "" {
		gpkgFile = createTmpFile()
		downloadGeopackage(gpkgFile, *gpkgURLParam)
	} else {
		var err error
		gpkgFile, err = os.Open(*gpkgPathParam)
		if err != nil {
			log.Fatal("Error opening file: ", gpkgPathParam)
		}
	}
	return gpkgFile
}

// Download a Geopackage and store it in a file
func downloadGeopackage(gpkgFile *os.File, url string) {
	log.Printf("Starting download for: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("Download failed: didn't get a 200 statuscode, instead got %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	_, err = io.Copy(gpkgFile, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Geopackage downloaded")
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
		log.Fatal("Error with querying Geopackage: ", errDb)
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
		log.Fatal("Error with querying Geopackage: ", errDb)
	}
	columns, errDb := statement.Columns()
	if errDb != nil {
		log.Fatal("Error with querying Geopackage: ", errDb)
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
		log.Fatal("Error with querying Geopackage: ", errDb)
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
	if _, err := os.Stat("output"); os.IsNotExist(err) {
		os.Mkdir("output", 0777)
	}
	fileName := "output/" + layer + ".html"
	errFile := ioutil.WriteFile(fileName, htmlBuffer.Bytes(), 0777)
	if errFile != nil {
		log.Fatal("Cannot create html file", errFile)
	}
}

// Remove created temporary file
func cleanup(gpkgFile *os.File, gpkgURLParam *string) {
	if *gpkgURLParam != "" {
		tempFileName := gpkgFile.Name()
		err := os.Remove(tempFileName)
		if err != nil {
			log.Print("Could not cleanup temp file: ", err)
		} else {
			log.Print("Deleted temporary file: ", tempFileName)
		}
	}
}

// Program finished with succes, log time and exit
func programFinishedSuccesfully(startTime time.Time) {
	log.Println("Finished in (" + strconv.FormatFloat(time.Now().Sub(startTime).Seconds(), 'f', 2, 64) + "s).")
	os.Exit(0)
}

const htmlStart = "<!-- MapServer Template -->\n<html>\n\t<head>\n\t\t<title>GetFeatureInfo output</title>\n\t</head>\n\t<style type=\"text/css\">table.featureInfo, table.featureInfo td, table.featureInfo th { border: 1px solid #ddd; border-collapse: collapse; margin: 0; padding: 0; font-size: 90%; padding: .2em .1em; } table.featureInfo th { padding: .2em .2em; font-weight: bold; background: #eee; } table.featureInfo td { background: #fff; } table.featureInfo tr.odd td { background: #eee; } table.featureInfo caption { text-align: left; font-size: 100%; font-weight: bold; padding: .2em .2em; }</style>\n\t<body>\n\t\t<table class=\"featureInfo\">\n"
const htmlLayer = "\t\t\t<caption class=\"featureInfo\">{{.layer}}</caption>\n\t\t\t<tr>\n"
const htmlColumnHead = "\t\t\t\t<th>{{.column}}</th>\n"
const htmlColumnRow = "\t\t\t\t<td>[{{.column}}]</td>\n"
const htmlEnd = "\t\t\t</tr>\n\t\t</table>\n\t</body>\n</html>\n<!-- Generated by PDOK ( https://www.pdok.nl/ ) -->"
