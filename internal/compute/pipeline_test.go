package compute

import (
	"strings"
	"testing"

	"pii/internal/compute/detect"
	"pii/internal/compute/mask"
)

func newTestPipeline() *Pipeline {
	ner := detect.NewNERDetector(4000, 200)
	ner.Preload()
	return NewPipeline(
		detect.CreateStructuralRegistry(),
		ner,
		detect.NewContextRule(200),
		mask.NewMasker("partial"),
	)
}

func TestPipelineBasic(t *testing.T) {
	p := newTestPipeline()
	masked, types := p.Process("Клиент Иванов Иван Иванович, паспорт 4509 123456")
	if !strings.Contains(masked, "И. И. И.") {
		t.Fatalf("expected fio masked, got %q", masked)
	}
	if !strings.Contains(masked, "45** ****56") {
		t.Fatalf("expected passport masked, got %q", masked)
	}
	if !contains(types, "fio") || !contains(types, "passport") {
		t.Fatalf("expected fio and passport types, got %v", types)
	}
}

func TestPipelineNoPII(t *testing.T) {
	p := newTestPipeline()
	masked, _ := p.Process("Привет, мир!")
	if masked != "Привет, мир!" {
		t.Fatalf("expected unchanged, got %q", masked)
	}
}

func TestPipelinePushkinNoContext(t *testing.T) {
	p := newTestPipeline()
	masked, _ := p.Process("Пушкин")
	if masked != "Пушкин" {
		t.Fatalf("expected unchanged, got %q", masked)
	}
}

func TestPipelinePushkinWithPassport(t *testing.T) {
	p := newTestPipeline()
	masked, _ := p.Process("Пушкин Александр Сергеевич, паспорт 4509 123456")
	if !strings.Contains(masked, "П. А. С.") {
		t.Fatalf("expected fio masked, got %q", masked)
	}
}

func TestPipelinePINWithoutCard(t *testing.T) {
	p := newTestPipeline()
	masked, _ := p.Process("пин-код 1234")
	if strings.Contains(masked, "***") {
		t.Fatalf("expected pin not masked, got %q", masked)
	}
}

func TestPipelinePINWithCard(t *testing.T) {
	p := newTestPipeline()
	masked, _ := p.Process("карта 4111 1111 1111 1111, пин-код 1234")
	if !strings.Contains(masked, "4111 **** **** 1111") {
		t.Fatalf("expected card masked, got %q", masked)
	}
}

func TestPipelineBareDateNotPII(t *testing.T) {
	p := newTestPipeline()
	masked, _ := p.Process("встреча 12.01.2024")
	if strings.Contains(masked, "**") {
		t.Fatalf("expected date not masked, got %q", masked)
	}
}

func TestPipelineBirthDate(t *testing.T) {
	p := newTestPipeline()
	masked, _ := p.Process("дата рождения 12.01.1990")
	if !strings.Contains(masked, "**.**.1990") {
		t.Fatalf("expected date masked, got %q", masked)
	}
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
