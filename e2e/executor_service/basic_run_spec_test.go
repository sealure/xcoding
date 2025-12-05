//go:build e2e

package executor_e2e

import (
	"fmt"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Basic pipeline run creates job/step and succeeds", func() {
	println("s1")
	var token string
	ginkgo.BeforeEach(func() {
		Expect(ping()).To(BeTrue())
		t, err := adminLogin()
		Expect(err).NotTo(HaveOccurred())
		token = t
	})

	println("s2")
	ginkgo.It("creates pipeline, starts build, runs a job with two steps", func() {
		suf := uniqueNano()
		pid, err := createProject(fmt.Sprintf("ci_exec_proj_%d", suf), "executor e2e", false, token)
		Expect(err).NotTo(HaveOccurred())

		yaml := strings.Join([]string{
			"jobs:",
			"  build:",
			"    container: alpine:3.19",
			"    - name: step1",
			"      run: echo hello1",
			"    - name: step2",
			"      run: echo hello2",
		}, "\n")
		println("s3")

		plID, err := createPipeline(token, pid, fmt.Sprintf("pipeline_%d", suf), "basic run", yaml)
		Expect(err).NotTo(HaveOccurred())
		println("s4")
		bID, err := startBuild(token, plID, "main", "e2e-basic")
		println("s4.1")
		Expect(err).NotTo(HaveOccurred())
		println("s4.2")

		Eventually(func() string {
			st, _ := getBuildStatus(token, bID)
			return st
		}, 60*time.Second, 1*time.Second).Should(Equal("BUILD_STATUS_SUCCEEDED"))
		println("s5")
		lines, err := getExecutorLogs(token, bID)
		Expect(err).NotTo(HaveOccurred())
		hasStep1 := false
		hasStep2 := false
		println("s6")
		for _, ln := range lines {
			if strings.Contains(ln, "__step_begin__ step1") {
				hasStep1 = true
			}
			if strings.Contains(ln, "__step_end__ step1") {
				hasStep1 = true
			}
			if strings.Contains(ln, "__step_begin__ step2") {
				hasStep2 = true
			}
			if strings.Contains(ln, "__step_end__ step2") {
				hasStep2 = true
			}
		}
		println("ssssss")
		Expect(hasStep1 && hasStep2).To(BeTrue(), "missing step markers in logs")
	})
})
