package cloudgrpc

import (
	"bufio"
	"bytes"
	"cloud/pkg/cloudv1"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log/slog"
	"os"
)

type Client struct {
	api cloudv1.CloudClient
	log *slog.Logger
}

// New creates grpc client.
func New(
	addr string,
	log *slog.Logger,
) (*Client, error) {
	const fn = "cloudgrpc.New"

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("%v: %w", fn, err)
	}

	client := cloudv1.NewCloudClient(conn)

	return &Client{
		api: client,
		log: log,
	}, nil
}

// Upload uploads image from cloud.
func (c *Client) Upload(src string) error {
	const fn = "cloudgrpc.Upload"
	stream, err := c.api.Upload(context.Background())
	if err != nil {
		c.log.Error(err.Error(), slog.String("fn", fn))
		return fmt.Errorf("%s: %w", fn, err)
	}

	// try to open source file
	file, err := os.Open(src)
	if err != nil {
		c.log.Error(err.Error(), slog.String("fn", fn))
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer file.Close()

	r := bufio.NewReader(file)

	// send filename
	data := &cloudv1.UploadRequest{
		Data: &cloudv1.UploadRequest_Name{
			Name: src,
		},
	}
	if err := stream.Send(data); err != nil {
		c.log.Error(err.Error(), slog.String("fn", fn))
		return fmt.Errorf("%s: %w", fn, err)
	}

	// send file
	// recommended chunk size for streamed messages appears to be 16-64KiB
	buf := make([]byte, 64*1024)
	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			c.log.Error(err.Error(), slog.String("fn", fn))
			return fmt.Errorf("%s: %w", fn, err)
		}

		data := &cloudv1.UploadRequest{
			Data: &cloudv1.UploadRequest_Chunk{
				Chunk: buf[:n],
			},
		}
		err = stream.Send(data)
		if err == io.EOF {
			break
		}
		if err != nil {
			c.log.Error(err.Error(), slog.String("fn", fn))
			return fmt.Errorf("%s: %w", fn, err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		c.log.Error(err.Error(), slog.String("fn", fn))
		return fmt.Errorf("%s: %w", fn, err)
	}

	c.log.Info("successful upload", slog.String("fn", fn), slog.String("resp", resp.String()))

	return nil
}

// Download downloads image from cloud.
func (c *Client) Download(path string, filename string) error {
	const fn = "cloudgrpc.Download"

	stream, err := c.api.Download(context.Background(), &cloudv1.DownloadRequest{Name: filename})
	if err != nil {
		c.log.Error(err.Error(), slog.String("fn", fn))
		return fmt.Errorf("%s: %w", fn, err)
	}

	buf := bytes.Buffer{}

	size := 0
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.log.Error(err.Error(), slog.String("fn", fn))
			return fmt.Errorf("%s: %w", fn, err)
		}

		chunk := data.GetChunk()
		size += len(chunk)

		_, err = buf.Write(chunk)
		if err != nil {
			c.log.Error(err.Error(), slog.String("fn", fn))
			return fmt.Errorf("%s: %w", fn, err)
		}
	}

	file, err := os.Create(path + filename)
	if err != nil {
		c.log.Error(err.Error(), slog.String("fn", fn))
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer file.Close()

	_, err = buf.WriteTo(file)
	if err != nil {
		c.log.Error(err.Error(), slog.String("fn", fn))
		return fmt.Errorf("%s: cannot write buf to file: %w", fn, err)
	}

	c.log.Info("successful download", slog.String("fn", fn), slog.Int("size", size))

	return nil
}

// List prints images on cloud.
func (c *Client) List() error {
	const fn = "cloudgrpc.List"

	resp, err := c.api.List(context.Background(), &cloudv1.ListRequest{})
	if err != nil {
		c.log.Error(err.Error(), slog.String("fn", fn))
		return fmt.Errorf("%s: %w", fn, err)
	}

	fmt.Println("Name | Created at | Updated at")
	for _, val := range resp.Files {
		fmt.Printf("%s | %s | %s\n", val.Name, val.CreatedAt, val.UpdatedAt)
	}

	return nil
}
