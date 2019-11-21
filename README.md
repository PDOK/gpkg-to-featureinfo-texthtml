# gpkg-to-featureinfo-texthtml

![GitHub license](https://img.shields.io/github/license/PDOK/gpkg-to-featureinfo-texthtml.svg)
![GitHub release](https://img.shields.io/github/release/PDOK/gpkg-to-featureinfo-texthtml.svg)
![Go report](https://goreportcard.com/badge/github.com/pdok/gpkg-to-featureinfo-texthtml.svg)

<img src="gpkg-logo.PNG" alt="gpkg-to-featureinfo-texthtml logo" width="400px" title="logo"/>
Generate HTML feature-info templates for Mapserver from GPKG

## Build
To build the application, make sure you have GoLang (v1.11+) installed.  
`go build`

## Test
To test the application, make sure you have GoLang (v1.11+) installed.  
`go test`

## Usage with sources
You can use either an URL where a Geopackage can be downloaded or use a local Geopackage.

Example with an URL:  
`go run main.go -gpkg-url https://domain.nl/geopackages/dataset/1/dataset.gpkg`

Example with a local file:  
`go run main.go -gpkg-path /home/user/downloads/afvalwater.gpkg`

## Usage with binary (Linux)
You can use either an URL where a Geopackage can be downloaded or use a local Geopackage.

Example with an URL:  
`gpkg-to-featureinfo-texthtml -gpkg-url https://domain.nl/geopackages/dataset/1/dataset.gpkg`

Example with a local file:  
`gpkg-to-featureinfo-texthtml -gpkg-path /home/user/downloads/afvalwater.gpkg`

## Usage with binary (Windows)
You can use either an URL where a Geopackage can be downloaded or use a local Geopackage.

Example with an URL:  
`gpkg-to-featureinfo-texthtml.exe -gpkg-url https://domain.nl/geopackages/dataset/1/dataset.gpkg`

Example with a local file:  
`gpkg-to-featureinfo-texthtml.exe -gpkg-path C:\files\afvalwater.gpkg`
