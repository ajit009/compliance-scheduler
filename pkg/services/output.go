package services

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"aws-compliance-scheduler/pkg/config"
	"aws-compliance-scheduler/pkg/templates"
)

// OutputHolder holds key-value pairs that belong together in the output
type OutputHolder struct {
	Contents map[string]string
}

// OutputArray holds all the different OutputHolders that will be provided as
// output, as well as the keys (headers) that will actually need to be printed
type OutputArray struct {
	Title    string
	SortKey  string
	Contents []OutputHolder
	Keys     []string
}

// GetContentsMap returns a stringmap of the output contents
func (output OutputArray) GetContentsMap() []map[string]string {
	total := make([]map[string]string, 0, len(output.Contents))
	for _, holder := range output.Contents {
		values := make(map[string]string)
		for _, key := range output.Keys {
			if val, ok := holder.Contents[key]; ok {
				values[key] = val
			}
		}
		total = append(total, values)
	}
	return total
}

// Write will provide the output as configured in the configuration
func (output OutputArray) Write(settings config.Config) {
	switch settings.GetOutputFormat() {
	case "csv":
		output.toCSV(*settings.OutputFile)
	case "html":
		output.toHTML(*settings.OutputFile, settings.ShouldAppend())
	default:
		output.toJSON(*settings.OutputFile)
	}
}

func (output OutputArray) toCSV(outputFile string) {
	total := [][]string{}
	total = append(total, output.Keys)
	for _, holder := range output.Contents {
		values := make([]string, len(output.Keys))
		for counter, key := range output.Keys {
			if val, ok := holder.Contents[key]; ok {
				values[counter] = val
			}
		}
		total = append(total, values)
	}
	var target io.Writer
	if outputFile == "" {
		target = os.Stdout
	} else {
		file, err := os.Create(outputFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		target = bufio.NewWriter(file)
	}
	w := csv.NewWriter(target)

	for _, record := range total {
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}

func (output OutputArray) toJSON(outputFile string) {
	jsonString, _ := json.Marshal(output.GetContentsMap())

	err := PrintByteSlice(jsonString, outputFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	
}

func (output OutputArray) toHTML(outputFile string, append bool) {
	t := template.New("table")
	t, _ = t.Parse(templates.HTMLTableTemplate)
	var baseTemplate string
	if append {
		originalfile, err := ioutil.ReadFile(outputFile)
		if err != nil {
			panic(err)
		}
		baseTemplate = string(originalfile)
	} else {
		b := template.New("base")
		b, _ = b.Parse(templates.BaseHTMLTemplate)
		baseBuf := new(bytes.Buffer)
		b.Execute(baseBuf, output)
		baseTemplate = baseBuf.String()
	}
	tableBuf := new(bytes.Buffer)
	t.Execute(tableBuf, output)
	resultString := strings.Replace(baseTemplate, "<div id='end'></div>", tableBuf.String(), 1)
	err := PrintByteSlice([]byte(resultString), outputFile)
	if err != nil {
		log.Fatal(err.Error())
	}
}

// PrintByteSlice prints the provided contents to stdout or the provided filepath
func PrintByteSlice(contents []byte, outputFile string) error {
	var target io.Writer
	var err error
	if outputFile == "" {
		target = os.Stdout
	} else {
		target, err = os.Create(outputFile)
		if err != nil {
			return err
		}
	}
	w := bufio.NewWriter(target)
	w.Write(contents)
	err = w.Flush()
	return err
}

// AddHolder adds the provided OutputHolder to the OutputArray
func (output *OutputArray) AddHolder(holder OutputHolder) {
	var contents []OutputHolder
	if output.Contents != nil {
		contents = output.Contents
	}
	contents = append(contents, holder)
	if output.SortKey != "" {
		sort.Slice(contents,
			func(i, j int) bool {
				return contents[i].Contents[output.SortKey] < contents[j].Contents[output.SortKey]
			})
	}
	output.Contents = contents
}
