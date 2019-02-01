package gofs

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"strings"
)

// DefaultEncoders contains the default list of encoders per MIME type.
var DefaultEncoders = EncoderGroup{
	"xml":        EncoderMakerFunc(func(w io.Writer) Encoder { return &xmlEncoder{w} }),
	"json":       EncoderMakerFunc(func(w io.Writer) Encoder { return &jsonEncoder{w, false} }),
	"prettyjson": EncoderMakerFunc(func(w io.Writer) Encoder { return &jsonEncoder{w, true} }),
	"yaml":       EncoderMakerFunc(func(w io.Writer) Encoder { return &yamlEncoder{w} }),
}

type (
	// An Encoder encodes data from v.
	Encoder interface {
		Encode(v interface{}) error
	}

	// An EncoderGroup maps MIME types to EncoderMakers.
	EncoderGroup map[string]EncoderMaker

	// An EncoderMaker creates and returns a new Encoder.
	EncoderMaker interface {
		NewEncoder(w io.Writer) Encoder
	}

	// EncoderMakerFunc is an adapter for creating EncoderMakers
	// from functions.
	EncoderMakerFunc func(w io.Writer) Encoder
)

// NewEncoder implements the EncoderMaker interface.
func (f EncoderMakerFunc) NewEncoder(w io.Writer) Encoder {
	return f(w)
}

type xmlEncoder struct {
	w io.Writer
}

func (xe *xmlEncoder) Encode(v interface{}) error {
	xe.w.Write([]byte(xml.Header))
	defer xe.w.Write([]byte("\n"))
	e := xml.NewEncoder(xe.w)
	e.Indent("", "\t")
	return e.Encode(v)
}

type jsonEncoder struct {
	w      io.Writer
	pretty bool
}

func (je *jsonEncoder) Encode(v interface{}) error {
	if je.pretty {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		var out bytes.Buffer
		err = json.Indent(&out, b, "", "\t")
		if err != nil {
			return err
		}
		_, err = io.Copy(je.w, &out)
		if err != nil {
			return err
		}
		_, err = je.w.Write([]byte("\n"))
		return err
	}
	return json.NewEncoder(je.w).Encode(v)
}

type yamlEncoder struct {
	w io.Writer
}

func (ye *yamlEncoder) Encode(v interface{}) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	_, err = ye.w.Write(b)
	return err
}

func PrettyJsonString(v interface{}) string {
	output, _ := json.MarshalIndent(v, "", "  ")
	return string(output)
}

// DefaultDecoders contains the default list of decoders per MIME type.
var DefaultDecoders = DecoderGroup{
	"xml":  DecoderMakerFunc(func(r io.Reader) Decoder { return xml.NewDecoder(r) }),
	"json": DecoderMakerFunc(func(r io.Reader) Decoder { return json.NewDecoder(r) }),
	"yaml": DecoderMakerFunc(func(r io.Reader) Decoder { return &yamlDecoder{r} }),
}

type (
	// A Decoder decodes data into v.
	Decoder interface {
		Decode(v interface{}) error
	}

	// A DecoderGroup maps MIME types to DecoderMakers.
	DecoderGroup map[string]DecoderMaker

	// A DecoderMaker creates and returns a new Decoder.
	DecoderMaker interface {
		NewDecoder(r io.Reader) Decoder
	}

	// DecoderMakerFunc is an adapter for creating DecoderMakers
	// from functions.
	DecoderMakerFunc func(r io.Reader) Decoder
)

// NewDecoder implements the DecoderMaker interface.
func (f DecoderMakerFunc) NewDecoder(r io.Reader) Decoder {
	return f(r)
}

type yamlDecoder struct {
	r io.Reader
}

func (yd *yamlDecoder) Decode(v interface{}) error {
	b, err := ioutil.ReadAll(yd.r)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, v)
}

func ReadAsCSV(val string) ([]string, error) {
	if val == "" {
		return []string{}, nil
	}
	stringReader := strings.NewReader(val)
	csvReader := csv.NewReader(stringReader)
	return csvReader.Read()
}
