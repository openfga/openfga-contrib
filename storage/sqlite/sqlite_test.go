package sqlite

import (
	"testing"

	"github.com/openfga/openfga/pkg/storage/test"
)

func TestOpenFGA_StorageConformance(t *testing.T) {
	ds := New()
	test.RunAllTests(t, ds)
}
