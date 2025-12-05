//go:build e2e
// +build e2e

package code_repository_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCodeRepository(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Code Repository E2E Suite")
}