package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/cisco/senml"
	"io/ioutil"
	"net/http"
	"os"
)

var doIndentPtr = flag.Bool("i", false, "indent output")
var doPrintPtr = flag.Bool("print", false, "print output to stdout")
var doExpandPtr = flag.Bool("expand", false, "expand SenML records")
var postUrl = flag.String("post", "", "URL to HTTP POST output to")
var topic = flag.String("topic", "senml", "Apache Kafka topic or InfluxDB series name ")

var doJsonPtr = flag.Bool("json", false, "output JSON formatted SenML ")
var doCborPtr = flag.Bool("cbor", false, "output CBOR formatted SenML ")
var doXmlPtr = flag.Bool("xml", false, "output XML formatted SenML ")
var doCsvPtr = flag.Bool("csv", false, "output CSV formatted SenML ")
var doMpackPtr = flag.Bool("mpack", false, "output MessagePack formatted SenML ")
var doLinpPtr = flag.Bool("linp", false, "output InfluxDB LineProtcol formatted SenML ")

var doIJsonStreamPtr = flag.Bool("ijson", false, "input JSON formatted SenML")
//var doIJsonLinePtr = flag.Bool("ijsonl", false, "input JSON formatted SenML lines")
var doIXmlPtr = flag.Bool("ixml", false, "input XML formatted SenML ")
var doICborPtr = flag.Bool("icbor", false, "input CBOR formatted SenML ")
var doIMpackPtr = flag.Bool("impack", false, "input MessagePack formatted SenML ")

func decodeTimed(msg []byte) (senml.SenML, error) {
	var s senml.SenML
	var err error

	var format senml.Format = senml.JSON
	switch {
	case *doIJsonStreamPtr:
		format = senml.JSON
	//case *doIJsonLinePtr:
	//	format = senml.JSON
	case *doICborPtr:
		format = senml.CBOR
	case *doIXmlPtr:
		format = senml.XML
	case *doIMpackPtr:
		format = senml.MPACK
	}

	s, err = senml.Decode(msg, format)

	return s, err
}

func outputData(data []byte) error {
	if *doPrintPtr {
		fmt.Print(string(data))
	}

	if len(*postUrl) != 0 {
		fmt.Println("PostURL=<" + string(*postUrl) + ">")
		buffer := bytes.NewBuffer(data)
		_, err := http.Post(string(*postUrl), "application/senml+json", buffer)
		if err != nil {
			fmt.Println("Post to", string(*postUrl), " got error", err.Error())
			return err
		}
	}

	return nil
}

func processData(dataIn []byte) error {
	var s senml.SenML
	var err error

	s, err = decodeTimed(dataIn)
	if err != nil {
		fmt.Println("Decode of SenML failed")
		return err
	}

	//fmt.Println( "Senml:", senml.Records )
	if *doExpandPtr {
		s = senml.Normalize(s)
	}

	var dataOut []byte
	options := senml.OutputOptions{}
	if *doIndentPtr {
		options.PrettyPrint = *doIndentPtr
	}
	options.Topic = string(*topic)
	var format senml.Format = senml.JSON
	switch {
	case *doJsonPtr:
		format = senml.JSON
	case *doCborPtr:
		format = senml.CBOR
	case *doXmlPtr:
		format = senml.XML
	case *doCsvPtr:
		format = senml.CSV
	case *doMpackPtr:
		format = senml.MPACK
	case *doLinpPtr:
		format = senml.LINEP
	}
	dataOut, err = senml.Encode(s, format, options)
	if err != nil {
		fmt.Println("Encode of SenML failed")
		return err
	}

	err = outputData(dataOut)
	if err != nil {
		fmt.Println("Output of SenML failed:", err)
		return err
	}

	return nil
}

func main() {
	var err error

	flag.Parse()

	// load the input
	msg, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Println("error reading SenML file", err)
		os.Exit(1)
	}

	err = processData(msg)
}