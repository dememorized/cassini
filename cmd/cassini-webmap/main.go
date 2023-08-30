package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	_ "embed"
)

//go:embed map.html
var index []byte

func main() {
	tiles := flag.String("tiles", "out", "Path to the directory containing tile-files")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/tiles/", func(writer http.ResponseWriter, request *http.Request) {
		if !strings.HasSuffix(request.URL.Path, ".png") {
			writer.WriteHeader(http.StatusNotFound)
			writer.Write([]byte("Not found"))
			return
		}

		tp, err := readTilePath(*tiles, request.URL.Path)
		if err != nil {
			fmt.Println(err)
			return
		}

		f, err := os.Open(tp.String())
		if err != nil {
			fmt.Println(err)
			return
		}

		io.Copy(writer, f)
	})
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write(index)
		if err != nil {
			panic(err)
		}
	})

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		panic(err)
	}
}

type TilePath struct {
	Base   string
	Zoom   int
	X      int
	Y      int
	Suffix string
}

func (t TilePath) String() string {
	return filepath.Join(t.Base, strconv.Itoa(t.Zoom), strconv.Itoa(t.X), strconv.Itoa(t.Y)+t.Suffix)
}

var validSuffixes = map[string]struct{}{".png": {}}

func readTilePath(base, p string) (TilePath, error) {
	tp := TilePath{
		Base:   base,
		Zoom:   0,
		X:      0,
		Y:      0,
		Suffix: "",
	}

	rest, y := path.Split(p)
	rest, x := path.Split(strings.TrimSuffix(rest, "/"))
	_, zoom := path.Split(strings.TrimSuffix(rest, "/"))

	suffix := strings.ToLower(filepath.Ext(y))
	if _, ok := validSuffixes[suffix]; !ok {
		return TilePath{}, fmt.Errorf("invalid filetype: %s", suffix)
	}
	tp.Suffix = suffix

	y = strings.SplitN(y, ".", 2)[0]

	var errY, errX, errZoom error
	tp.Y, errY = strconv.Atoi(y)
	tp.X, errX = strconv.Atoi(x)
	tp.Zoom, errZoom = strconv.Atoi(zoom)
	if err := errors.Join(errY, errX, errZoom); err != nil {
		return TilePath{}, err
	}

	return tp, nil
}
