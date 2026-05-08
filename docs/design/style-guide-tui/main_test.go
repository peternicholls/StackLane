package main

import (
	"strings"
	"testing"
)

func TestGuideContainsCoreSections(t *testing.T) {
	styleGuide := buildGuide()
	wantSections := []string{"start", "intent", "colour", "components", "report", "guided", "assistance", "separation", "review"}

	for _, sectionID := range wantSections {
		if _, ok := styleGuide.sectionByID(sectionID); !ok {
			t.Fatalf("missing guide section %q", sectionID)
		}
	}
}

func TestPlainGuideContainsDesignContract(t *testing.T) {
	styleGuide := buildGuide()
	var builder strings.Builder
	renderPlain(&builder, styleGuide, "")
	plain := builder.String()

	for _, want := range []string{
		"This is the terminal design style guide",
		"Colour carries semantic meaning only",
		"Run this project",
		"Help me fix these",
		"Bubble Tea and Lip Gloss are implementation tools",
		"Report surfaces remain useful",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("plain guide missing %q:\n%s", want, plain)
		}
	}
}

func TestGuidedExampleUsesMoreEllipsis(t *testing.T) {
	styleGuide := buildGuide()
	section, ok := styleGuide.sectionByID("guided")
	if !ok {
		t.Fatal("missing guided section")
	}
	rendered := section.Render(78)

	if !strings.Contains(rendered, "More…") {
		t.Fatalf("guided section should use product More… label:\n%s", rendered)
	}
	if strings.Contains(rendered, "More...") {
		t.Fatalf("guided section should not use three-dot More label:\n%s", rendered)
	}
}
