package virtualmachine

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPoolManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "virtualmachine webhook Suite")
}
