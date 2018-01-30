package main

import (
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"

    "github.com/linkedin/goavro"
)

func bail(err error) {
    fmt.Fprintf(os.Stderr, "%s\n", err)
    os.Exit(1)
}

func usage() {
    executable, err := os.Executable()
    if err != nil {
        executable = os.Args[0]
    }
    base := filepath.Base(executable)
    fmt.Fprintf(os.Stderr, "Usage of %s:\n", base)
    fmt.Fprintf(os.Stderr, "\t%s [-compression null|deflate|snappy] schema.avsc output.avro\n", base)
    flag.PrintDefaults()
    os.Exit(2)
}

func main() {
    compressionName := flag.String("compression", "null", "compression codec ('null', 'deflate', 'snappy'; default: 'null')")
    flag.Parse()

    schemaBytes, err := ioutil.ReadFile(flag.Arg(0))
    if err != nil {
        bail(err)
    }
    
    codec, err := goavro.NewCodec(string(schemaBytes))
    if err != nil {
        bail(err)
    }

    textual := []byte(`{"name": "x", "favorite_number":{"float": 3.0}, "favorite_color":{"string":"z"}}`)

    // Convert textual Avro data (in Avro JSON format) to native Go form
    native, _, err := codec.NativeFromTextual(textual)
    if err != nil {
        fmt.Println(err)
    }

    // Convert native Go form to binary Avro data
    binary, err := codec.BinaryFromNative(nil, native)
    if err != nil {
        fmt.Println(err)
    }

    fh, err := os.Create("flag.Arg(1)")
    if err != nil {
        bail(err)
    }
    defer func(ioc io.Closer) {
        if err := ioc.Close(); err != nil {
            bail(err)
        }
    }(fh)

    ocfw, err := goavro.NewOCFWriter(goavro.OCFConfig{
        W:               fh,
        Codec:           codec,
        CompressionName: *compressionName,
    })
    if err != nil {
        bail(err)
    }

    var datum interface{}

    for len(binary) > 0 {
        datum, binary, err = codec.NativeFromBinary(binary)
        if err != nil {
            if err == io.EOF {
                err = nil
                break
            }
            bail(err)
        }
        if err = ocfw.Append([]interface{}{datum}); err != nil {
            bail(err)
        }
    }

    if err != nil {
        bail(err)
    }

}