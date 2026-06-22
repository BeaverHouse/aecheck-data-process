package external

import (
	"aecheck-data-process/internal/constants"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/types"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"

	"github.com/BeaverHouse/go-common/env"
	"github.com/BeaverHouse/go-common/logger"
	"github.com/HugoSmits86/nativewebp"
	"golang.org/x/image/webp"
)

var imageBasePath = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("failed to get user home directory: " + err.Error())
	}
	return home + "\\Pictures\\aecheck"
}()

func getStyleSuffix(category types.AECategory, style types.AEStyle) string {
	if category == types.AECategoryColab {
		return ""
	}
	return constants.StyleSuffix[string(style)]
}

func buildImageFileName(gameID int, styleSuffix string, stellar bool) string {
	if stellar {
		return fmt.Sprintf("%d%s%s.png", gameID, styleSuffix, constants.StellarSuffix)
	}
	return fmt.Sprintf("%d%s.png", gameID, styleSuffix)
}

func getUploadPath(baseDir string, dryrun bool) string {
	if dryrun {
		return "test/" + baseDir
	}
	return baseDir
}

func uploadFile(path string, fileName string, data []byte) error {
	oracleUploadURL := env.GetEnv("AECHECK_UPLOAD_URL", "")
	fullURL := fmt.Sprintf("%s/o/aecheck/%s/%s", oracleUploadURL, path, fileName)
	common.Log.Info("Uploading file", logger.Field{Key: "url", Value: fullURL})

	req, err := http.NewRequest(http.MethodPut, fullURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("uploadFile: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("uploadFile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("uploadFile: status %d, body: %s", resp.StatusCode, string(body))
	}

	common.Log.Info("File uploaded", logger.Field{Key: "path", Value: path}, logger.Field{Key: "fileName", Value: fileName})
	return nil
}

func isWebP(imgBytes []byte) bool {
	return len(imgBytes) >= 4 && string(imgBytes[0:4]) == "RIFF"
}

func decodeImage(imgBytes []byte) (image.Image, error) {
	if isWebP(imgBytes) {
		return webp.Decode(bytes.NewReader(imgBytes))
	}
	return png.Decode(bytes.NewReader(imgBytes))
}

func convertToWebp(imgBytes []byte) ([]byte, error) {
	img, err := decodeImage(imgBytes)
	if err != nil {
		return nil, fmt.Errorf("convertToWebp: %w", err)
	}

	var webpBuf bytes.Buffer
	if err := nativewebp.Encode(&webpBuf, img, nil); err != nil {
		return nil, fmt.Errorf("convertToWebp: %w", err)
	}

	return webpBuf.Bytes(), nil
}

func uploadImagePair(uploadPath, characterID string, imgBytes []byte) error {
	var pngBytes, webpBytes []byte

	if isWebP(imgBytes) {
		img, err := decodeImage(imgBytes)
		if err != nil {
			return fmt.Errorf("uploadImagePair: %w", err)
		}

		var pngBuf bytes.Buffer
		if err := png.Encode(&pngBuf, img); err != nil {
			return fmt.Errorf("uploadImagePair: %w", err)
		}
		pngBytes = pngBuf.Bytes()
		webpBytes = imgBytes

		common.Log.Info("Converted WebP to PNG",
			logger.Field{Key: "webpSize", Value: len(imgBytes)},
			logger.Field{Key: "pngSize", Value: len(pngBytes)})
	} else {
		pngBytes = imgBytes

		var err error
		webpBytes, err = convertToWebp(imgBytes)
		if err != nil {
			return err
		}

		common.Log.Info("Converted PNG to WebP",
			logger.Field{Key: "pngSize", Value: len(pngBytes)},
			logger.Field{Key: "webpSize", Value: len(webpBytes)},
			logger.Field{Key: "reduction", Value: fmt.Sprintf("%.1f%%", 100*(1-float64(len(webpBytes))/float64(len(pngBytes))))})
	}

	if err := uploadFile(uploadPath, characterID+".png", pngBytes); err != nil {
		return err
	}

	if err := uploadFile(uploadPath, characterID+".webp", webpBytes); err != nil {
		return err
	}
	return nil
}

func UploadCharacterImage(info types.CharacterInfoFromAEWiki, id int, isFourStar bool, dryrun bool) error {
	characterID := fmt.Sprintf("char%04d", id)

	style := info.Style
	if isFourStar {
		style = types.StyleFOUR
	}

	styleSuffix := getStyleSuffix(info.Category, style)

	imageName := buildImageFileName(info.GameID, styleSuffix, false)
	imagePath := fmt.Sprintf("%s\\%s", imageBasePath, imageName)

	common.Log.Info("Reading character image", logger.Field{Key: "path", Value: imagePath})
	imgBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("UploadCharacterImage: %w", err)
	}

	uploadPath := getUploadPath("character", dryrun)
	if err := uploadImagePair(uploadPath, characterID, imgBytes); err != nil {
		return err
	}
	common.Log.Info("Character image uploaded", logger.Field{Key: "source", Value: imageName})

	if isFourStar {
		return nil
	}

	stellarImageName := buildImageFileName(info.GameID, styleSuffix, true)
	stellarImagePath := fmt.Sprintf("%s\\%s", imageBasePath, stellarImageName)

	common.Log.Info("Reading stellar image", logger.Field{Key: "path", Value: stellarImagePath})
	stellarImageBytes, err := os.ReadFile(stellarImagePath)
	if err != nil {
		common.Log.Warn("Stellar image not found", logger.Field{Key: "path", Value: stellarImageName})
		return nil
	}

	stellarUploadPath := getUploadPath("staralign", dryrun)
	if err := uploadImagePair(stellarUploadPath, characterID, stellarImageBytes); err != nil {
		return err
	}
	common.Log.Info("Stellar image uploaded", logger.Field{Key: "source", Value: stellarImageName})

	return nil
}

func UploadBuddyImage(gameID int, buddyID string, dryrun bool) error {
	imageName := fmt.Sprintf("%d.png", gameID)
	imagePath := fmt.Sprintf("%s\\%s", imageBasePath, imageName)

	common.Log.Info("Reading buddy image", logger.Field{Key: "path", Value: imagePath})
	imgBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("UploadBuddyImage: %w", err)
	}

	uploadPath := getUploadPath("buddy", dryrun)
	if err := uploadImagePair(uploadPath, buddyID, imgBytes); err != nil {
		return err
	}
	common.Log.Info("Buddy image uploaded", logger.Field{Key: "source", Value: imageName})
	return nil
}
