package libgograbber

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

func buildResponseHeader(header *http.Response) string {
	buf := new(bytes.Buffer)
	header.Write(buf)
	return buf.String()
}

func Report(s *State, targets chan Host) ([]string, error) {
	var results []Host
	for host := range targets {
		results = append(results, host)
	}

	var reportFiles []string
	for _, format := range s.OutputFormats {
		format = strings.TrimSpace(strings.ToLower(format))
		var reportFile string
		var err error
		switch format {
		case "md":
			reportFile, err = MarkdownReport(s, results)
		case "json":
			reportFile, err = JsonReport(s, results)
		case "csv":
			reportFile, err = CsvReport(s, results)
		case "xml":
			reportFile, err = XmlReport(s, results)
		}
		if err != nil {
			return reportFiles, err
		}
		if reportFile != "" {
			reportFiles = append(reportFiles, reportFile)
		}
	}
	return reportFiles, nil
}

func MarkdownReport(s *State, results []Host) (string, error) {
	var report bytes.Buffer
	currTime := GetTimeString()

	var reportFile string
	if s.ProjectName != "" {
		reportFile = path.Join(s.ReportDirectory, fmt.Sprintf("%v_%v_Report.md", SanitiseFilename(s.ProjectName), currTime))
	} else {
		reportFile = path.Join(s.ReportDirectory, fmt.Sprintf("%v_Report.md", currTime))
	}
	file, err := os.Create(reportFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Header
	report.WriteString(fmt.Sprintf("# Gograbber report - %v (%v)\n", s.ProjectName, currTime))
	for _, URLComponent := range results {
		url := fmt.Sprintf("%v://%v:%v/%v\n", URLComponent.Protocol, URLComponent.HostAddr, URLComponent.Port, URLComponent.Path)
		report.WriteString(fmt.Sprintf("## %v\n", url))
		if URLComponent.HTTPResp != nil {
			report.WriteString("### Response Headers\n")
			report.WriteString(fmt.Sprintf("```\n%v```\n", buildResponseHeader(URLComponent.HTTPResp)))
			report.WriteString("### Response Body File\n")
			if URLComponent.ResponseBodyFilename != "" {
				report.WriteString(fmt.Sprintf("\n`%v`\n", URLComponent.ResponseBodyFilename))
			} else {
				report.WriteString(fmt.Sprintf("\n`<No output file>`\n"))
			}
		}
		report.WriteString("### Screenshot\n")
		report.WriteString(fmt.Sprintf("![%v](../../%v)\n", URLComponent.ScreenshotFilename, URLComponent.ScreenshotFilename))

		file.WriteString(report.String())
		report.Reset()
	}
	return reportFile, nil
}

func JsonReport(s *State, results []Host) (string, error) {
	currTime := GetTimeString()
	var reportFile string
	if s.ProjectName != "" {
		reportFile = path.Join(s.ReportDirectory, fmt.Sprintf("%v_%v_Report.json", SanitiseFilename(s.ProjectName), currTime))
	} else {
		reportFile = path.Join(s.ReportDirectory, fmt.Sprintf("%v_Report.json", currTime))
	}
	file, err := os.Create(reportFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(results)
	if err != nil {
		return "", err
	}
	return reportFile, nil
}

func CsvReport(s *State, results []Host) (string, error) {
	currTime := GetTimeString()
	var reportFile string
	if s.ProjectName != "" {
		reportFile = path.Join(s.ReportDirectory, fmt.Sprintf("%v_%v_Report.csv", SanitiseFilename(s.ProjectName), currTime))
	} else {
		reportFile = path.Join(s.ReportDirectory, fmt.Sprintf("%v_Report.csv", currTime))
	}
	file, err := os.Create(reportFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Protocol", "Host", "Port", "Path", "Screenshot", "ResponseFile"})
	for _, host := range results {
		writer.Write([]string{
			host.Protocol,
			host.HostAddr,
			strconv.Itoa(host.Port),
			host.Path,
			host.ScreenshotFilename,
			host.ResponseBodyFilename,
		})
	}
	return reportFile, nil
}

type XmlReportWrapper struct {
	XMLName xml.Name `xml:"gograbber_report"`
	Project string   `xml:"project,attr"`
	Hosts   []Host   `xml:"host"`
}

func XmlReport(s *State, results []Host) (string, error) {
	currTime := GetTimeString()
	var reportFile string
	if s.ProjectName != "" {
		reportFile = path.Join(s.ReportDirectory, fmt.Sprintf("%v_%v_Report.xml", SanitiseFilename(s.ProjectName), currTime))
	} else {
		reportFile = path.Join(s.ReportDirectory, fmt.Sprintf("%v_Report.xml", currTime))
	}
	file, err := os.Create(reportFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	wrapper := XmlReportWrapper{
		Project: s.ProjectName,
		Hosts:   results,
	}

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	file.WriteString(xml.Header)
	err = encoder.Encode(wrapper)
	if err != nil {
		return "", err
	}
	return reportFile, nil
}

func SanitiseFilename(UnsanitisedFilename string) string {
	r := regexp.MustCompile("[^0-9a-zA-Z-._]")
	return r.ReplaceAllString(UnsanitisedFilename, "_")
}

func writerWorker(l Loggers, writeChan chan []byte, filename string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if os.IsNotExist(err) {
		file, err = os.Create(filename)
	}
	if err != nil {
		l.Error.Printf("Failed to open output file %s: %v\n", filename, err)
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for b := range writeChan {
		if len(b) > 0 {
			writer.Write(b)
			writer.Flush()
		}
	}
}
