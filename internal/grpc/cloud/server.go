package cloud

import (
	"bufio"
	"bytes"
	"cloud/internal/config"
	"cloud/internal/storage"
	"cloud/internal/storage/drive"
	"cloud/pkg/cloudv1"
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type Cloud interface {
	Upload(filename string, buf bytes.Buffer) error
	CanUpload(filename string) (bool, error)
	List() ([]drive.Image, error)
	Search(filename string) (*os.File, error)
}

type Server struct {
	cloudv1.UnimplementedCloudServer
	cloud     Cloud
	log       *slog.Logger
	cfg       config.CloudConfig
	limitUD   chan struct{}
	limitList chan struct{}
}

func New(
	cloud Cloud,
	log *slog.Logger,
	cfg config.CloudConfig,
) *Server {
	return &Server{
		cloud:     cloud,
		log:       log,
		cfg:       cfg,
		limitUD:   make(chan struct{}, cfg.LimitUD),
		limitList: make(chan struct{}, cfg.LimitList),
	}
}

func Register(gRPC *grpc.Server, server *Server) {
	cloudv1.RegisterCloudServer(gRPC, server)
}

// Upload save image on storage.
func (s *Server) Upload(stream cloudv1.Cloud_UploadServer) error {
	const fn = "cloud.Upload"

	s.limitUD <- struct{}{}
	defer func() {
		<-s.limitUD
	}()

	s.log.Info("upload/download clients", slog.String("fn", fn), slog.Int("current", len(s.limitUD)),
		slog.Int("max", cap(s.limitUD)))

	// get filename
	req, err := stream.Recv()
	if err != nil {
		s.log.Error(err.Error(), slog.String("fn", fn))
		return status.Errorf(codes.Internal, ErrInternal.Error())
	}

	// check errors
	filename := filepath.Base(req.GetName())
	if filename == "" {
		s.log.Info(ErrEmptyFilename.Error(), slog.String("fn", fn))
		return status.Error(codes.InvalidArgument, ErrEmptyFilename.Error())
	}
	ext := filepath.Ext(filename)
	if _, ok := s.cfg.AvailableExt[ext]; !ok {
		err = &ErrImageExt{s.cfg.AvailableExt}
		s.log.Info(err.Error(), slog.String("fn", fn))
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// checking whether we can upload the file to the server
	can, err := s.cloud.CanUpload(filename)
	if err != nil {
		s.log.Info(err.Error(), slog.String("fn", fn))
		return status.Error(codes.Internal, ErrInternal.Error())
	}
	if !can {
		s.log.Info(storage.ErrFileExists.Error(), slog.String("fn", fn))
		return status.Error(codes.AlreadyExists, storage.ErrFileExists.Error())
	}

	buf := bytes.Buffer{}
	size := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.log.Error(err.Error(), slog.String("fn", fn))
			return status.Errorf(codes.Internal, ErrInternal.Error())
		}

		chunk := req.GetChunk()
		size += len(chunk)

		if size > s.cfg.MaxImageSize {
			err = &ErrImageMaxSize{maxImageSize: s.cfg.MaxImageSize}
			s.log.Info(err.Error(), slog.String("fn", fn))
			return status.Errorf(codes.InvalidArgument, err.Error())
		}

		_, err = buf.Write(chunk)
		if err != nil {
			s.log.Error(err.Error(), slog.String("fn", fn))
			return status.Errorf(codes.Internal, ErrInternal.Error())
		}
	}

	// call service layer
	err = s.cloud.Upload(filename, buf)
	if err != nil {
		s.log.Error(err.Error(), slog.String("fn", fn))
		if errors.Is(err, storage.ErrFileExists) {
			return status.Errorf(codes.AlreadyExists, err.Error())
		}
		return status.Errorf(codes.Internal, ErrInternal.Error())
	}

	err = stream.SendAndClose(&cloudv1.UploadResponse{
		Name: filename,
		Size: uint32(size),
	})

	if err != nil {
		s.log.Error(err.Error(), slog.String("fn", fn))
		return status.Errorf(codes.Internal, ErrInternal.Error())
	}

	s.log.Info("file uploaded", slog.String("fn", fn), slog.String("filename", filename))

	return nil
}

// List returns list of images.
func (s *Server) List(context.Context, *cloudv1.ListRequest) (*cloudv1.ListResponse, error) {
	const fn = "cloud.List"

	s.limitList <- struct{}{}
	defer func() {
		<-s.limitList
	}()

	s.log.Info("images list clients", slog.String("fn", fn), slog.Int("current",
		len(s.limitList)), slog.Int("max", cap(s.limitList)))

	images, err := s.cloud.List()
	if err != nil {
		s.log.Info(err.Error(), slog.String("fn", fn))
		return nil, status.Errorf(codes.Internal, ErrInternal.Error())
	}

	res := make([]*cloudv1.FileStructure, 0, len(images))
	for _, image := range images {
		file := &cloudv1.FileStructure{
			Name:      image.Name,
			CreatedAt: image.CreatedAt,
			UpdatedAt: image.UpdatedAt,
		}
		res = append(res, file)
	}

	return &cloudv1.ListResponse{
		Files: res,
	}, nil
}

// Download downloads image from storage.
func (s *Server) Download(req *cloudv1.DownloadRequest, stream cloudv1.Cloud_DownloadServer) error {
	const fn = "cloud.Download"

	s.limitUD <- struct{}{}
	defer func() {
		<-s.limitUD
	}()

	s.log.Info("upload/download clients", slog.String("fn", fn), slog.Int("current",
		len(s.limitUD)), slog.Int("max", cap(s.limitUD)))

	filename := req.GetName()
	file, err := s.cloud.Search(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return status.Errorf(codes.NotFound, ErrNotExist.Error())
		}
		s.log.Info(err.Error(), slog.String("fn", fn))
		return status.Errorf(codes.Internal, ErrInternal.Error())
	}
	defer file.Close()

	r := bufio.NewReader(file)

	// recommended chunk size for streamed messages appears to be 16-64KiB
	chunk := make([]byte, 64*1024)
	for {
		n, err := r.Read(chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.log.Error(err.Error(), slog.String("fn", fn))
			return status.Errorf(codes.Internal, ErrInternal.Error())
		}

		data := &cloudv1.DownloadResponse{
			Chunk: chunk[:n],
		}

		if err := stream.Send(data); err != nil {
			s.log.Error(err.Error(), slog.String("fn", fn))
			return status.Errorf(codes.Internal, ErrInternal.Error())
		}
	}

	s.log.Info("file downloaded", slog.String("fn", fn), slog.String("filename", filename))

	return nil
}
