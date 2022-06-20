package config

import (
	"aws-compliance-scheduler/pkg/log"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
)

// CSVFileToMap  reads csv file into slice of map
// slice is the line number
// map[string]string where key is column name
func CSVFileToMap(filePath string) (returnMap []map[string]string, err error) {
	// read csv file
	csvfile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	defer csvfile.Close()

	reader := csv.NewReader(csvfile)

	rawCSVdata, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	header := []string{} // holds first row (header)
	for lineNum, record := range rawCSVdata {

		// for first row, build the header slice
		if lineNum == 0 {
			for i := 0; i < len(record); i++ {
				header = append(header, strings.TrimSpace(record[i]))
			}
		} else {
			// for each cell, map[string]string k=header v=value
			line := map[string]string{}
			for i := 0; i < len(record); i++ {
				line[header[i]] = record[i]
			}
			returnMap = append(returnMap, line)
		}
	}

	return
}

/*
Prepares map for a specific resource type.
Assumes a hardcoded structure of CSV with columns :
 - enviroment
 - team
 - name [ aws resource name]
 - resource [aws resource like rds, s3 etc]
 - skip [ should scheduler ignore]
*/
func ReadResourceData(filePath string, sheetName string, res string) (returnMap map[string]map[string]string, err error) {
	log.Info("Reading data from " + filePath + " and using data on " + sheetName + " for " + res)
	awsdata := make(map[string]map[string]string)
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Error("config.ReadResourceData::Error::" + err.Error())
		return
	}
	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Error("config.ReadResourceData::Error::" + err.Error())
	}
	for _, row := range rows {
		log.Tracef("config.ReadResourceData::INFO::current row::%v\n", row)
		if row[4] == res {
			rowData := make(map[string]string)
			rowData["environment"] = row[1]
			rowData["team"] = row[2]
			rowData["name"] = row[0]
			rowData["resource"] = row[3]
			rowData["skip"] = row[4]
			awsdata[row[0]] = rowData
		} else {
			log.Info("config.ReadResourceData::INFO::skipping row as " + row[4] + " is not " + res)
		}

	}
	log.Info("config.ReadResourceData::INFO::completed reading xlsx")
	return awsdata, err
}

/*
* returns data from the excel sheet awsdata stored locally
 */

func ReadAWSData(filePath string, sheetName string) (returnMap map[string]map[string]string, err error) {
	fmt.Println("Reading data from " + filePath + " and using data on " + sheetName)
	awsdata := make(map[string]map[string]string)
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		fmt.Println("config.ReadAWSData::Error::" + err.Error())
		return
	}
	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Println("config.ReadAWSData::Error::" + err.Error())
	}
	for _, row := range rows {
		// fmt.Print(row)
		rowData := make(map[string]string)
		rowData["environment"] = row[1]
		rowData["team"] = row[2]
		rowData["repo"] = row[]
		rowData["name"] = row[0]
		rowData["resource"] = row[3]
		rowData["skip"] = row[4]
		awsdata[row[0]] = rowData

	}
	return awsdata, err
}

func ReadAWSDataFromS3() (returnMap map[string]map[string]string, err error) {

	return
}
