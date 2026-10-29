package models

import (
	"testing"

	"gorm.io/gorm"
)

type longTextFakeDialector struct {
	gorm.Dialector
	name string
}

func (d longTextFakeDialector) Name() string { return d.name }

func TestLongTextGormDBDataType(t *testing.T) {
	cases := map[string]string{
		"mysql":    "mediumtext",
		"postgres": "text",
		"sqlite":   "mediumtext",
	}
	for driver, want := range cases {
		db := &gorm.DB{Config: &gorm.Config{Dialector: longTextFakeDialector{name: driver}}}
		if got := LongText("").GormDBDataType(db, nil); got != want {
			t.Fatalf("%s LongText type=%q want %q", driver, got, want)
		}
	}
}

func TestLongTextPtrHelpers(t *testing.T) {
	if got := NewLongTextPtr("abc"); got == nil || string(*got) != "abc" {
		t.Fatalf("NewLongTextPtr=%v", got)
	}
	if got := LongTextFromPtr(nil); got != nil {
		t.Fatalf("LongTextFromPtr(nil)=%v", got)
	}
	raw := "xyz"
	if got := LongTextFromPtr(&raw); got == nil || string(*got) != raw {
		t.Fatalf("LongTextFromPtr=%v", got)
	}
}
