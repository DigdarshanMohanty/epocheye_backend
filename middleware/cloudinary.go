package middleware

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var CLD *cloudinary.Cloudinary

func InitCloudinary() error {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return fmt.Errorf("Cloudinary env vars missing")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return err
	}

	CLD = cld
	return nil
}

func UploadFile(ctx context.Context, fileBytes []byte, fileName string) (string, error) {
	if CLD == nil {
		return "", fmt.Errorf("Cloudinary client not initialized")
	}

	reader := bytes.NewReader(fileBytes)
	resp, err := CLD.Upload.Upload(ctx, reader, uploader.UploadParams{
		PublicID: fileName,
		Folder:   "user_avatars",
	})
	if err != nil {
		return "", fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return resp.SecureURL, nil
}
