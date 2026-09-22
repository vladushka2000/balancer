# Трек 2 — Go detect/ner + context + merge

## Статус: DONE (проверено 2026-09-22)

Трек 2 уже реализован в рамках общего Go-перевода. Проверка приёмки:

## Файлы

- `internal/compute/detect/ner.go` — `NERDetector`: regex+dict (ФИО/адрес/орган),
  чанкование `chunkChars` с перекрытием `overlap`, обработка чанков в goroutines,
  сдвиг спанов на offset чанка, дедупликация по `(start, end, type)`.
- `internal/compute/detect/context.go` — `ContextRule.Filter`: словари
  famous/org_addresses/markers + контекстное окно `window`; структурные спаны
  (source="regex") не фильтруются.
- `internal/compute/detect/merge.go` — `MergeSpans`: сортировка по start,
  разрешение пересечений (prefer regex → длиннее), ПИН-гейт (pin без card → отброс).
- Тесты: `ner_test.go` (3), `context_test.go` (5), `merge_test.go` (6) — всего 14.

## Критерии приёмки

- [x] `ner.go` — regex+dict NER, чанки, goroutines, offset, dedup.
- [x] `context.go` — фильтр по словарям + контекстное окно.
- [x] `merge.go` — пересечения + ПИН-гейт.
- [x] Все 14 тестов зелёные (`go test ./internal/compute/detect/...`).
- [x] `go build ./...`, `go vet ./...`, `gofmt -l .` — чисто.
- [x] Комментарии — только пакетные `//`-доки (допустимы по §0B); закомментированного
      кода нет.

## Замечание

Словари лежат в `internal/compute/detect/dicts/*.txt` (famous ≥20, org_addresses ≥10,
markers ≥13, months 12) и подключаются через `//go:embed`. `months.txt` используется
структурным детектором дат (трек 1), не NER.