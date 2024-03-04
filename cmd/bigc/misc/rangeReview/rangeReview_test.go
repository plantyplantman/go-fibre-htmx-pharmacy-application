package main

import (
	"strings"
	"testing"

	"github.com/plantyplantman/bcapi/pkg/product"
	"gorm.io/gorm"
)

func TestIsDeletedMinfos(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "Test with backslash",
			input: "\\test",
			want:  true,
		},
		{
			name:  "Test with hash",
			input: "#test",
			want:  true,
		},
		{
			name:  "Test with exclamation mark",
			input: "!test",
			want:  true,
		},
		{
			name:  "Test without special character",
			input: "test",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDeletedMinfos(tt.input); got != tt.want {
				t.Errorf("isDeletedMinfos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHi(t *testing.T) {
	service, err := product.NewDefaultService()
	if err != nil {
		t.Fatal(err)
	}

	ps, err := service.FetchProducts(func(d *gorm.DB) *gorm.DB {
		return d.Where("name LIKE ?", "%"+strings.ToUpper(strings.TrimSpace("DURO"))+"%")
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, p := range ps {
		t.Logf("%+v\n", p)
	}
}
