package runner

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/projectdiscovery/subfinder/pkg/resolve"
)

// OutPutter outputs content to writers.
type OutPutter struct {
	JSON bool
}

type jsonResult struct {
	Host   string `json:"host"`
	IP     string `json:"ip"`
	Source string `json:"source"`
}

// NewOutputter creates a new Outputter
func NewOutputter(json bool) *OutPutter {
	return &OutPutter{JSON: json}
}

func (o *OutPutter) createFile(filename, outputDirectory string, appendtoFile bool) (*os.File, error) {
	if filename == "" {
		return nil, errors.New("empty filename")
	}

	absFilePath := filename

	if outputDirectory != "" {
		if _, err := os.Stat(outputDirectory); os.IsNotExist(err) {
			err := os.MkdirAll(outputDirectory, os.ModePerm)
			if err != nil {
				return nil, err
			}
		}
		absFilePath = path.Join(outputDirectory, filename)
	}

	var file *os.File
	var err error
	if appendtoFile {
		file, err = os.OpenFile(absFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		file, err = os.Create(absFilePath)
	}
	if err != nil {
		return nil, err
	}

	return file, nil
}

// WriteForChaos prepares the buffer to upload to Chaos
func (o *OutPutter) WriteForChaos(results map[string]resolve.HostEntry, writer io.Writer) error {
	bufwriter := bufio.NewWriter(writer)
	sb := &strings.Builder{}

	for _, result := range results {
		sb.WriteString(result.Host)
		sb.WriteString("\n")

		_, err := bufwriter.WriteString(sb.String())
		if err != nil {
			bufwriter.Flush()
			return err
		}
		sb.Reset()
	}
	return bufwriter.Flush()
}

// WriteHostIP writes the output list of subdomain to an io.Writer
func (o *OutPutter) WriteHostIP(results map[string]resolve.Result, writer io.Writer) error {
	var err error
	if o.JSON {
		err = writeJSONHostIP(results, writer)
	} else {
		err = writePlainHostIP(results, writer)
	}
	return err
}

func writePlainHostIP(results map[string]resolve.Result, writer io.Writer) error {
	bufwriter := bufio.NewWriter(writer)
	sb := &strings.Builder{}

	for _, result := range results {
		sb.WriteString(result.Host)
		sb.WriteString(",")
		sb.WriteString(result.IP)
		sb.WriteString(",")
		sb.WriteString(result.Source)
		sb.WriteString("\n")

		_, err := bufwriter.WriteString(sb.String())
		if err != nil {
			bufwriter.Flush()
			return err
		}
		sb.Reset()
	}
	return bufwriter.Flush()
}

func writeJSONHostIP(results map[string]resolve.Result, writer io.Writer) error {
	encoder := jsoniter.NewEncoder(writer)

	var data jsonResult

	for _, result := range results {
		data.Host = result.Host
		data.IP = result.IP
		data.Source = result.Source

		err := encoder.Encode(&data)
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteHostNoWildcard writes the output list of subdomain with nW flag to an io.Writer
func (o *OutPutter) WriteHostNoWildcard(results map[string]resolve.Result, writer io.Writer) error {
	hosts := make(map[string]resolve.HostEntry)
	for host, result := range results {
		hosts[host] = resolve.HostEntry{Host: result.Host, Source: result.Source}
	}

	return o.WriteHost(hosts, writer)
}

// WriteHost writes the output list of subdomain to an io.Writer
func (o *OutPutter) WriteHost(results map[string]resolve.HostEntry, writer io.Writer) error {
	var err error
	if o.JSON {
		err = writeJSONHost(results, writer)
	} else {
		err = writePlainHost(results, writer)
	}
	return err
}

func writePlainHost(results map[string]resolve.HostEntry, writer io.Writer) error {
	bufwriter := bufio.NewWriter(writer)
	sb := &strings.Builder{}

	for _, result := range results {
		sb.WriteString(result.Host)
		sb.WriteString("\n")

		_, err := bufwriter.WriteString(sb.String())
		if err != nil {
			bufwriter.Flush()
			return err
		}
		sb.Reset()
	}
	return bufwriter.Flush()
}

func writeJSONHost(results map[string]resolve.HostEntry, writer io.Writer) error {
	encoder := jsoniter.NewEncoder(writer)

	for _, result := range results {
		err := encoder.Encode(result)
		if err != nil {
			return err
		}
	}
	return nil
}
