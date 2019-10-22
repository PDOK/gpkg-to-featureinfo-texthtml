package main

import (
	"testing"
)

func Test_checkParameters(t *testing.T) {
	var testArgUrl *string
	temp := "-gpkgurl http://csu338.cs.kadaster.nl:8080/geopackages/afvalwater2016/1/afvalwater.gpkg"
	testArgUrl = &temp
	var testArgPath *string
	temp2 := ""
	testArgPath = &temp2
	checkParameters(testArgUrl, testArgPath)
	temp = ""
	temp2 = "-gpkgpath ./test.gpkg"
	testArgUrl = &temp
	testArgPath = &temp2
	checkParameters(testArgUrl, testArgPath)
}

func Test_createTmpFile(t *testing.T) {
	testFile := createTmpFile()
	if testFile == nil {
		t.Error("Could not create temporary file.")
	}
}

func Test_downloadGeopackage(t *testing.T) {
	testFile := createTmpFile()
	if testFile == nil {
		t.Error("Could not create temporary file.")
	}
	err := downloadGeopackage(testFile, "http://csu338.cs.kadaster.nl:8080/geopackages/afvalwater2016/1/afvalwater.gpkg")
	if err != nil {
		t.Error(err)
	}
}

func Test_generateHTMLForLayer(t *testing.T) {
	const expectedResult = "<!-- MapServer Template -->\n<html>\n\t<head>\n\t\t<title>GetFeatureInfo output</title>\n\t</head>\n\t<style type=\"text/css\">table.featureInfo, table.featureInfo td, table.featureInfo th { border: 1px solid #ddd; border-collapse: collapse; margin: 0; padding: 0; font-size: 90%; padding: .2em .1em; } table.featureInfo th { padding: .2em .2em; font-weight: bold; background: #eee; } table.featureInfo td { background: #fff; } table.featureInfo tr.odd td { background: #eee; } table.featureInfo caption { text-align: left; font-size: 100%; font-weight: bold; padding: .2em .2em; }</style>\n\t<body>\n\t\t<table class=\"featureInfo\">\n\t\t\t<caption class=\"featureInfo\">testLayer</caption>\n\t\t\t<tr>\n\t\t\t\t<th>testColumn1</th>\n\t\t\t\t<th>testColumn2</th>\n\t\t\t</tr>\n\t\t\t<tr>\n\t\t\t\t<td>[testColumn1]</td>\n\t\t\t\t<td>[testColumn2]</td>\n\t\t\t</tr>\n\t\t</table>\n\t</body>\n</html>\n<!-- Generated by PDOK ( https://www.pdok.nl/ ) -->"
	testColumn := []string{"testColumn1", "testColumn2", "geom", "shape_len"}
	geomColumn := []string{"geo"}
	htmlBuffer := generateHTMLForLayer("testLayer", testColumn, geomColumn)
	if htmlBuffer == nil {
		t.Error("No HTML was generated")
	}
	if htmlBuffer.String() != expectedResult {
		t.Errorf("Result was not OK.\nResult:\n%s.\nExpected:\n%s.", htmlBuffer.String(), expectedResult)
	}
}

func Test_checkColumn(t *testing.T) {
	badColumns := []string{"geom", "shape_len", "shape_leng", "shape_area", "Shape_Area", "geo"}
	geomColumn := []string{"geo"}
	for _, badColumn := range badColumns {
		if checkColumn(badColumn, geomColumn) {
			t.Errorf("%s should not be an valid column name.", badColumn)
		}
	}
}
