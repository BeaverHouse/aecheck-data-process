package main

import (
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/logic/parse"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if logic.IsLocalEnv() {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Failed to load .env file: %v", err)
		}
	}
	common.InitLogger()

	// 공지사항 ID
	announceID := 644 // 여기에 원하는 ID를 입력하세요

	languages := []string{"ko", "ja", "en"}

	for _, lang := range languages {
		fmt.Printf("\n=== Language: %s ===\n", lang)
		content, err := parse.GetHTMLcontent(announceID, lang)
		if err != nil {
			fmt.Printf("Error for %s: %v\n", lang, err)
			continue
		}
		fmt.Printf("Content for %s:\n%s\n", lang, content)
	}
}
