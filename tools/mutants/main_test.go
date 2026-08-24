package main

import (
	"reflect"
	"testing"
	"time"
)

func TestTestArgsSerializesTaggedIntegrationPackages(t *testing.T) {
	original := testTags
	t.Cleanup(func() { testTags = original })

	testTags = ""
	if got, want := testArgs("./..."), []string{"test", "./..."}; !reflect.DeepEqual(got, want) {
		t.Fatalf("untagged args=%v want=%v", got, want)
	}
	if got := testTimeout(); got != 10*time.Second {
		t.Fatalf("untagged timeout=%s", got)
	}

	testTags = "integration"
	if got, want := testArgs("./..."), []string{"test", "-p=1", "-tags", "integration", "./..."}; !reflect.DeepEqual(got, want) {
		t.Fatalf("tagged args=%v want=%v", got, want)
	}
	if got := testTimeout(); got != 30*time.Second {
		t.Fatalf("tagged timeout=%s", got)
	}
}
