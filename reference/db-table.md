# 데이터베이스 명세

AE Check에서 사용하는 데이터베이스 (이하 DB) 명세입니다.

## DB 정보

- DB Type: PostgreSQL
- Schema: aecheck

## 테이블 목록

### buddies

동료 정보를 저장합니다.

| 컬럼명       | 타입         | NULL     | 기본값            | 설명                                                   |
| ------------ | ------------ | -------- | ----------------- | ------------------------------------------------------ |
| buddy_id     | varchar(10)  | NOT NULL | -                 | 버디 ID                                                |
| character_id | varchar(10)  | NULL     | -                 | 버디와 대응되는 캐릭터 ID                              |
| get_path     | varchar(20)  | NULL     | -                 | 획득 경로 (대응되는 캐릭터가 있다면 NULL, 아니면 키값) |
| seesaa_url   | varchar(500) | NULL     | -                 | Seesaa Wiki URL                                        |
| aewiki_url   | varchar(500) | NULL     | -                 | AE Wiki URL                                            |
| created_at   | timestamp    | NOT NULL | CURRENT_TIMESTAMP | 생성일                                                 |
| updated_at   | timestamp    | NULL     | -                 | 수정일                                                 |
| deleted_at   | timestamp    | NULL     | -                 | 삭제일                                                 |

**제약조건**

- PRIMARY KEY (buddy_id)

### characters

캐릭터 정보를 저장합니다.

| 컬럼명          | 타입         | NULL     | 기본값            | 설명                                    |
| --------------- | ------------ | -------- | ----------------- | --------------------------------------- |
| character_id    | varchar(20)  | NOT NULL | -                 | 캐릭터 ID                               |
| character_code  | varchar(20)  | NOT NULL | -                 | 캐릭터 코드                             |
| category        | varchar(20)  | NOT NULL | -                 | 카테고리 (ENCOUNTER, FREE, COLAB)       |
| style           | varchar(10)  | NOT NULL | -                 | 스타일 (☆4, NS, AS, ES)                 |
| light_shadow    | varchar(10)  | NOT NULL | -                 | 천이면 light, 명이면 shadow             |
| max_manifest    | int4         | NOT NULL | -                 | 최대 현현 (0: 없음, 1: 현현, 2: 진현현) |
| is_awaken       | bool         | NOT NULL | -                 | 성도각성 여부                           |
| is_alter        | bool         | NOT NULL | -                 | 이시층 캐릭터 여부                      |
| alter_character | varchar(20)  | NULL     | -                 | 대응하는 이시층 캐릭터 ID               |
| seesaa_url      | varchar(500) | NULL     | -                 | Seesaa Wiki URL                         |
| aewiki_url      | varchar(500) | NULL     | -                 | AE Wiki URL                             |
| update_date     | date         | NULL     | -                 | 업데이트 날짜                           |
| created_at      | timestamp    | NOT NULL | CURRENT_TIMESTAMP | 생성일                                  |
| updated_at      | timestamp    | NULL     | -                 | 수정일                                  |
| deleted_at      | timestamp    | NULL     | -                 | 삭제일                                  |

**제약조건**

- PRIMARY KEY (character_id)

### dungeon_mappings

캐릭터와 던전 정보를 매핑합니다.

| 컬럼명       | 타입         | NULL     | 기본값                                             | 설명           |
| ------------ | ------------ | -------- | -------------------------------------------------- | -------------- |
| id           | int4         | NOT NULL | nextval('aecheck.ae_dungeon_map_id_seq'::regclass) | 매핑 ID (숫자) |
| character_id | varchar(20)  | NOT NULL | -                                                  | 캐릭터 ID      |
| dungeon_id   | varchar(20)  | NOT NULL | -                                                  | 던전 ID        |
| description  | varchar(500) | NULL     | -                                                  | 설명           |
| created_at   | timestamp    | NOT NULL | CURRENT_TIMESTAMP                                  | 생성일         |
| updated_at   | timestamp    | NULL     | -                                                  | 수정일         |
| deleted_at   | timestamp    | NULL     | -                                                  | 삭제일         |

**제약조건**

- PRIMARY KEY (id)

### dungeons

던전 정보를 저장합니다.

| 컬럼명     | 타입         | NULL     | 기본값            | 설명        |
| ---------- | ------------ | -------- | ----------------- | ----------- |
| dungeon_id | varchar(20)  | NOT NULL | -                 | 던전 ID     |
| altema_url | varchar(500) | NULL     | -                 | Altema URL  |
| aewiki_url | varchar(500) | NULL     | -                 | AE Wiki URL |
| created_at | timestamp    | NOT NULL | CURRENT_TIMESTAMP | 생성일      |
| updated_at | timestamp    | NULL     | -                 | 수정일      |
| deleted_at | timestamp    | NULL     | -                 | 삭제일      |

**제약조건**

- PRIMARY KEY (dungeon_id)

### personality_mappings

캐릭터와 퍼스널리티 정보를 매핑합니다.

| 컬럼명         | 타입         | NULL     | 기본값                                                 | 설명           |
| -------------- | ------------ | -------- | ------------------------------------------------------ | -------------- |
| id             | int4         | NOT NULL | nextval('aecheck.ae_personality_map_id_seq'::regclass) | 매핑 ID (숫자) |
| character_id   | varchar(20)  | NOT NULL | -                                                      | 캐릭터 ID      |
| personality_id | varchar(20)  | NOT NULL | -                                                      | 퍼스널리티 ID  |
| description    | varchar(500) | NULL     | -                                                      | 설명           |
| created_at     | timestamp    | NOT NULL | CURRENT_TIMESTAMP                                      | 생성일         |
| updated_at     | timestamp    | NULL     | -                                                      | 수정일         |
| deleted_at     | timestamp    | NULL     | -                                                      | 삭제일         |

**제약조건**

- PRIMARY KEY (id)

### translations

다국어 지원을 위한 번역 정보를 저장합니다.

| 컬럼명     | 타입         | NULL     | 기본값            | 설명               |
| ---------- | ------------ | -------- | ----------------- | ------------------ |
| key        | varchar(50)  | NOT NULL | -                 | i18n에 사용되는 키 |
| ko         | varchar(500) | NOT NULL | -                 | 한국어 번역        |
| en         | varchar(500) | NOT NULL | -                 | 영어 번역          |
| ja         | varchar(500) | NOT NULL | -                 | 일본어 번역        |
| created_at | timestamp    | NOT NULL | CURRENT_TIMESTAMP | 생성일             |
| updated_at | timestamp    | NULL     | -                 | 수정일             |
| deleted_at | timestamp    | NULL     | -                 | 삭제일             |

**제약조건**

- PRIMARY KEY (key)
