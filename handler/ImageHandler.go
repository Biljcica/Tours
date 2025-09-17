package handler

import (
	"io"
	"os"
	"path/filepath"
	"github.com/google/uuid"
	"mime"
	"net/http" 

	imagepb "database-example/proto/image"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ImageHandler implementira gRPC server za upload slika
type ImageHandler struct {
	imagepb.UnimplementedImageServiceServer
	UploadDir string // folder u kome cuvamo slike
}

func NewImageHandler(uploadDir string) *ImageHandler {
	return &ImageHandler{UploadDir: uploadDir}
}

func (h *ImageHandler) UploadImage(stream imagepb.ImageService_UploadImageServer) error {
	var info *imagepb.ImageInfo
	var file *os.File
	var path string
	var totalSize uint32

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			if info == nil {
				return status.Errorf(codes.InvalidArgument, "image info not received")
			}

			// Vrati response sa putanjom
			return stream.SendAndClose(&imagepb.UploadImageResponse{
				ImageId: uuid.New().String(),
				Path:    path,
				Size:    totalSize,
			})
		}
		if err != nil {
			return status.Errorf(codes.Internal, "stream error: %v", err)
		}

		switch x := req.RequestData.(type) {
		case *imagepb.UploadImageRequest_Info:
			info = x.Info

			// 1. Validacija ekstenzije
			ext := "." + info.ImageType
			/*allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
			if !allowed[ext] {
				return status.Errorf(codes.InvalidArgument, "unsupported file type: %s", ext)
			}*/

			// 2. Kreiraj UploadDir ako ne postoji
			if _, err := os.Stat(h.UploadDir); os.IsNotExist(err) {
				if err := os.MkdirAll(h.UploadDir, 0755); err != nil {
					return status.Errorf(codes.Internal, "failed to create upload dir: %v", err)
				}
			}

			// 3️. Jedinstveno ime fajla (uuid + originalno ime)
			filename := uuid.New().String() + "_" + info.Filename + ext
			path = filepath.Join(h.UploadDir, filename)

			// 4️. Otvori fajl za pisanje
			file, err = os.Create(path)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to create file: %v", err)
			}

		case *imagepb.UploadImageRequest_ChunkData:
			if file == nil {
				return status.Errorf(codes.FailedPrecondition, "file info must be sent before chunks")
			}
			n, writeErr := file.Write(x.ChunkData)
			if writeErr != nil {
				return status.Errorf(codes.Internal, "failed to write chunk: %v", writeErr)
			}
			totalSize += uint32(n)
		}
	}
}

func (h *ImageHandler) DownloadImage(req *imagepb.DownloadImageRequest, stream imagepb.ImageService_DownloadImageServer) error {
    path := filepath.Join(h.UploadDir, req.Filename)

    file, err := os.Open(path)
    if err != nil {
        return status.Errorf(codes.NotFound, "file not found: %v", err)
    }
    defer file.Close()

	// Dinamički MIME tip
    contentType := mime.TypeByExtension(filepath.Ext(req.Filename))
    if contentType == "" {
        // fallback: detektuj prema sadržaju fajla
        header := make([]byte, 512)
        n, _ := file.Read(header)
        contentType = http.DetectContentType(header[:n])
        file.Seek(0, io.SeekStart) // vrati pointer na početak fajla
    }

    buf := make([]byte, 32*1024) // 32KB chunk
    for {
        n, err := file.Read(buf)
        if n > 0 {
            if err := stream.Send(&imagepb.DownloadImageResponse{
                ChunkData:   buf[:n],
                ContentType: contentType, // ili dinamički preko filepath.Ext + mime.TypeByExtension
            }); err != nil {
                return err
            }
        }
        if err == io.EOF {
            break
        }
        if err != nil {
            return status.Errorf(codes.Internal, "failed to read file: %v", err)
        }
    }
    return nil
}

