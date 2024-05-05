package params

import "flag"

type Params struct {
	Addr     string
	Src      string
	Dest     string
	Filename string
	Method   string
}

func New() *Params {
	addr := flag.String("a", "localhost:44044", "the address to connect to")
	src := flag.String("src", "./images/client/test.png", "the source image path")
	dest := flag.String("dest", "./images/client/", "path for download images")
	filename := flag.String("fname", "", "download image with this filename from server")
	method := flag.String("m", "list", "grpc api method")

	flag.Parse()

	return &Params{
		Addr:     *addr,
		Src:      *src,
		Dest:     *dest,
		Filename: *filename,
		Method:   *method,
	}
}
