package typesystem_test

import (
	"testing"

	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

func TestGIRIdentifier_Matches(t *testing.T) {
	tests := []struct {
		name     string
		id       typesystem.GIRIdentifier
		pattern  typesystem.GIRIdentifier
		expected bool
	}{
		{
			name:     "Exact match",
			id:       typesystem.GIRIdentifier{"Button", "clicked", typesystem.GIRKindSignal},
			pattern:  typesystem.GIRIdentifier{"Button", "clicked", typesystem.GIRKindSignal},
			expected: true,
		},
		{
			name:     "Match with wildcard name",
			id:       typesystem.GIRIdentifier{"Button", "clicked", typesystem.GIRKindSignal},
			pattern:  typesystem.GIRIdentifier{"Button", "*", typesystem.GIRKindSignal},
			expected: true,
		},
		{
			name:     "Match with wildcard parent",
			id:       typesystem.GIRIdentifier{"Button", "clicked", typesystem.GIRKindSignal},
			pattern:  typesystem.GIRIdentifier{"*", "clicked", typesystem.GIRKindSignal},
			expected: true,
		},
		{
			name:     "Match with both wildcards",
			id:       typesystem.GIRIdentifier{"Button", "clicked", typesystem.GIRKindSignal},
			pattern:  typesystem.GIRIdentifier{"*", "*", typesystem.GIRKindSignal},
			expected: true,
		},
		{
			name:     "Kind mismatch",
			id:       typesystem.GIRIdentifier{"Button", "clicked", typesystem.GIRKindSignal},
			pattern:  typesystem.GIRIdentifier{"Button", "clicked", typesystem.GIRKindMethod},
			expected: false,
		},
		{
			name:     "Kind is Any (should match)",
			id:       typesystem.GIRIdentifier{"Button", "clicked", typesystem.GIRKindSignal},
			pattern:  typesystem.GIRIdentifier{"Button", "clicked", typesystem.GIRKindAny},
			expected: true,
		},
		{
			name:     "Pattern name mismatch",
			id:       typesystem.GIRIdentifier{"Button", "clicked", typesystem.GIRKindSignal},
			pattern:  typesystem.GIRIdentifier{"Button", "released", typesystem.GIRKindAny},
			expected: false,
		},
		{
			name:     "Pattern parent mismatch",
			id:       typesystem.GIRIdentifier{"Window", "clicked", typesystem.GIRKindSignal},
			pattern:  typesystem.GIRIdentifier{"Button", "clicked", typesystem.GIRKindAny},
			expected: false,
		},
		{
			name:     "No Parent match",
			id:       typesystem.GIRIdentifier{"", "clicked", typesystem.GIRKindFunction},
			pattern:  typesystem.GIRIdentifier{"", "clicked", typesystem.GIRKindAny},
			expected: true,
		},
		{
			name:     "No Parent mismatch",
			id:       typesystem.GIRIdentifier{"", "clicked", typesystem.GIRKindFunction},
			pattern:  typesystem.GIRIdentifier{"", "released", typesystem.GIRKindAny},
			expected: true,
		},
		{
			name:     "No Parent wildcard",
			id:       typesystem.GIRIdentifier{"", "clicked", typesystem.GIRKindFunction},
			pattern:  typesystem.GIRIdentifier{"", "*", typesystem.GIRKindAny},
			expected: true,
		},
		{
			name:     "Invalid Pattern",
			id:       typesystem.GIRIdentifier{"Parent", "clicked", typesystem.GIRKindFunction},
			pattern:  typesystem.GIRIdentifier{"Parent", "", typesystem.GIRKindAny},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Matches(tt.pattern); got != tt.expected {
				t.Errorf("Matches() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
