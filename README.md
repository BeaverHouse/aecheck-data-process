<p align="center">
  <a href="https://github.com/BeaverHouse/aecheck-data-process">
    <img src="logo.png" alt="Logo" width="100" height="100">
  </a>

  <p align="center">
    Data processing CLI for AE Check
  </p>

  <p align="center">
    <a href="https://go.dev/">
      <img src="https://img.shields.io/badge/Go-00ADD8.svg?style=flat&logo=Go&logoColor=white" alt="Go">
    </a>
    <a href="https://supabase.com/">
      <img src="https://img.shields.io/badge/Supabase-34b27b.svg?style=flat&logo=supabase&logoColor=white" alt="Supabase">
    </a>
    <a href="https://openai.com/codex/">
      <img src="https://img.shields.io/badge/Codex-black?logo=openaigym&logoColor=white" alt="Codex">
    </a>
    <a href="https://www.microsoft.com/en-us/windows">
      <img src="https://img.shields.io/badge/Windows Only-0078D4?logo=windows&logoColor=white" alt="Windows only">
    </a>
    <a href="./LICENSE">
      <img src="https://img.shields.io/github/license/BeaverHouse/aecheck-data-process" alt="License">
    </a>
  </p>
</p>

<!-- Content -->

<br>

## Description

[Go](https://go.dev/) CLI to update the data of the [AE Check](https://aecheck.com/) website.  
Used [Cobra](https://github.com/spf13/cobra) for CLI framework.

## History

| **Period**        | **Description**                                                            |
| :---------------- | :------------------------------------------------------------------------- |
| 2021.02 ~ 2022.09 | Manually processed the JSON data                                           |
| 2022.09 ~ 2023.09 | Used python script to process data (v2)                                    |
| 2023.09 ~ 2024.10 | Changed the data schema, configured auto-migration via GitHub Actions (v3) |
| 2024.10           | Migrated data to Supabase, deployed the backend (v3.1)                     |
| 2025.11           | Migrated data process code to Go                                           |
| 2026.02           | Add data processing for tier info (3 sources)                              |
| 2026.06           | Advanced automation for scraping wiki home & Windows scheduling            |

## Major data reference

| **Data source** | **Description**                                       |
| :-------------- | :---------------------------------------------------- |
| Raw application | I'm using mobile emulator to extract images directly. |
| [altema.jp]     | To scrape japanese data                               |
| [AE wiki]       | To scrape english data                                |
| [Seesaa Wiki]   | For additional information                            |
| [Another Tier]  | For tier reference                                    |

[AE Check]: https://aecheck.com/
[AE wiki]: https://anothereden.wiki/
[altema.jp]: https://altema.jp/anaden/
[Seesaa Wiki]: https://anothereden.game-info.wiki
[Another Tier]: https://anothertier.com/

## Limitation

- Because data process needs mobile emulator, this repository is only runnable in Windows.
- As of April 2026, [AE wiki] has enabled Cloudflare protection, blocking automated HTTP requests.
  So this CLI is semi-automatic; when a wiki URL is requested, a browser window opens; solve the Cloudflare
  challenge manually, then press Enter in the terminal to continue.

## Documentation

- [Database tables](./docs/db-table.md)

<br>

## Contributing

See the [CONTRIBUTING.md](./CONTRIBUTING.md).

<br>
<br>

## Attribution

Logo icon is extracted from Another Eden APK.
