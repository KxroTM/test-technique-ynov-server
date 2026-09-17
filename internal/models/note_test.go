package models

import "testing"

func TestNoteStatusIsValid(t *testing.T) {
	testCases := []struct {
		status NoteStatus
		valid  bool
	}{
		{StatusTodo, true},
		{StatusInProgress, true},
		{StatusDone, true},
		{NoteStatus("termine"), false},
		{NoteStatus("TODO"), false},
		{NoteStatus(""), false},
	}

	for _, testCase := range testCases {
		if got := testCase.status.IsValid(); got != testCase.valid {
			t.Errorf("NoteStatus(%q).IsValid() = %v, attendu %v",
				testCase.status, got, testCase.valid)
		}
	}
}

func TestNoteStatusLabel(t *testing.T) {
	testCases := map[NoteStatus]string{
		StatusTodo:       "Non fait",
		StatusInProgress: "En cours",
		StatusDone:       "Terminé",
	}

	for status, expected := range testCases {
		if got := status.Label(); got != expected {
			t.Errorf("NoteStatus(%q).Label() = %q, attendu %q", status, got, expected)
		}
	}
}
