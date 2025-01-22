# Obsidian Tools

## Contents

- Validate Obsidian Plugin Manifest

### Validate Obsidian Plugin Manifest

A tool to check an Obsidian plugin manifest against community rules as desribed in [Validate Plugin Entry workflow](https://github.com/obsidianmd/obsidian-releases/blob/master/.github/workflows/validate-plugin-entry.yml) of the obsidianmd/obsidian-releases project.

#### Run

`go run ./validate-obsidian-plugin-manifest`

#### Test

`go run ./validate-obsidian-plugin-manifest.go --manifest testdata/good-plugin-manifest.json`  
`go run ./validate-obsidian-plugin-manifest.go --manifest testdata/bad-plugin-manifest.json`
