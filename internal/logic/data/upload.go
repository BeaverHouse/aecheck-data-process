package data

import (
	"aecheck-data-process/internal/constants"
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/types"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"go.uber.org/zap"
)

// Uploads a file to the Oracle Object Storage.
func uploadFile(path string, fileName string, data []byte) error {
	oracleUploadURL := logic.GetEnv("AECHECK_UPLOAD_URL", "")

	req, err := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/o/aecheck/%s/%s", oracleUploadURL, path, fileName),
		bytes.NewReader(data))
	if err != nil {
		return common.WrapErrorWithContext("UploadFile", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return common.WrapErrorWithContext("UploadFile", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return common.WrapErrorWithContext("UploadFile", fmt.Errorf("failed to upload image: status %d, body: %s", resp.StatusCode, string(body)))
	}

	common.LogInfo("File uploaded", zap.String("path", path), zap.String("fileName", fileName))
	return nil
}

// Uploads the character image from Windows to the Oracle Object Storage.
func UploadCharacterImage(info types.CharacterInfoFromAEWiki, id int, isFourStar bool, dryrun bool) error {
	characterID := fmt.Sprintf("char%04d", id)

	if isFourStar {
		info.Style = types.StyleFOUR
	}

	imageBasePath := "C:\\Users\\LHU\\Documents\\XuanZhi9\\Pictures"
	imageName := fmt.Sprintf("%d%s.png", info.GameID, constants.IMG_SUFFIXS[string(info.Style)])

	fmt.Printf("Getting image from %s\\%s\n", imageBasePath, imageName)
	imgBytes, err := os.ReadFile(fmt.Sprintf("%s\\%s", imageBasePath, imageName))
	if err != nil {
		panic(err)
	}

	path := "character"
	if dryrun {
		path = "test/character"
	}

	err = uploadFile(path, fmt.Sprintf("%s.png", characterID), imgBytes)
	if err != nil {
		panic(err)
	}
	err = uploadFile(path, fmt.Sprintf("%s.webp", characterID), imgBytes)
	if err != nil {
		panic(err)
	}
	common.LogInfo("Character image uploaded", zap.String("path", fmt.Sprintf("%d%s.png", info.GameID, constants.IMG_SUFFIXS[string(info.Style)])))

	stellarPath := "staralign"
	if dryrun {
		stellarPath = "test/staralign"
	}

	stellarImageName := fmt.Sprintf("%d%s_opened.png", info.GameID, constants.IMG_SUFFIXS[string(info.Style)])

	fmt.Printf("Getting stellar image from %s\\%s\n", imageBasePath, stellarImageName)
	stellarImageBytes, err := os.ReadFile(fmt.Sprintf("%s\\%s", imageBasePath, stellarImageName))
	if err != nil {
		common.LogWarn("There's no stellar image", zap.String("path", stellarImageName))
		return nil
	} else {
		err = uploadFile(stellarPath, fmt.Sprintf("%s.png", characterID), stellarImageBytes)
		if err != nil {
			panic(err)
		}
		err = uploadFile(stellarPath, fmt.Sprintf("%s.webp", characterID), stellarImageBytes)
		if err != nil {
			panic(err)
		}
		common.LogInfo("Stellar image uploaded", zap.String("path", stellarImageName))
	}

	return nil
}
