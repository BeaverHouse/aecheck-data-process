---
name: update-tier
description: 지정된 폴더의 txt 파일 3개(altema.txt, seesaa.txt, anothertier.txt)를 읽고 tier.json을 생성한다. 티어 데이터 업데이트 시 사용.
argument-hint: "[폴더경로]"
disable-model-invocation: true
---

# Update Tier from Text Files

지정된 폴더의 txt 파일 3개를 읽고 tier.json을 생성한다.

폴더경로 미지정 시 현재 작업 디렉토리 사용: $ARGUMENTS

## Input Files (폴더 내)

- `altema.txt`: Altema 99점 캐릭터 (일본어)
- `seesaa.txt`: Seesaa Wiki EXC/SSS 캐릭터 (일본어)
- `anothertier.txt`: AnotherTier Best/SSS/SS 캐릭터 (영어)

## 파일 형식

```
https://... (URL - 무시)

캐릭터1
캐릭터2(AS)
캐릭터3(ES)
...
```

## 파싱 규칙

### 일본어 (altema.txt, seesaa.txt)
- `이름(AS)` → name=이름, style=AS
- `이름(ES)` → name=이름, style=ES
- `이름` → name=이름, style=NS
- lang="ja", is_alter=false

### 영어 (anothertier.txt)
- `Name AS` → name=Name, style=AS
- `Name ES` → name=Name, style=ES
- `Name` → name=Name, style=NS
- `Name AC` → name="Name (Alter)", style=NS, is_alter=true
- `Name AC AS` → name="Name (Alter)", style=AS, is_alter=true
- lang="en"

## Output (tier.json)

지정된 폴더에 생성. **Tier 계산 없이 각 사이트별 목록만 출력**:

```json
{
  "altema": [
    {"name": "アルマ", "style": "AS", "is_alter": false, "lang": "ja"}
  ],
  "seesaa": [
    {"name": "セスタ", "style": "AS", "is_alter": false, "lang": "ja"}
  ],
  "anothertier": [
    {"name": "Necoco", "style": "ES", "is_alter": false, "lang": "en"},
    {"name": "Myunfa (Alter)", "style": "AS", "is_alter": true, "lang": "en"}
  ]
}
```

## 실행

tier.json 생성 후 Go 스크립트로 DB 업데이트:

```bash
go run cmd/update_tier/main.go -json [폴더경로]/tier.json -dryrun
go run cmd/update_tier/main.go -json [폴더경로]/tier.json
```
