package flags_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/ofek/csi-gcs/pkg/flags"
)

var _ = Describe("Flags", func() {

	Describe("MergeFlags", func() {
		It("Should Merge", func() {
			Expect(
				MergeFlags(
					map[string]string{
						"bucket":   "test",
						"location": "US",
					},
					map[string]string{
						"bucket":    "test2",
						"projectId": "csi-gcs",
						"foo":       "bar",
					},
				),
			).To(Equal(map[string]string{
				"bucket":    "test2",
				"location":  "US",
				"projectId": "csi-gcs",
			}))
		})
	})
	Describe("MergeAnnotations", func() {
		It("Should Merge", func() {
			Expect(
				MergeAnnotations(
					map[string]string{
						"bucket":   "test",
						"location": "US",
					},
					map[string]string{
						"gcs.csi.ofek.dev/bucket":     "test2",
						"gcs.csi.ofek.dev/project-id": "csi-gcs",
						"gcs.csi.ofek.dev/foo":        "bar",
					},
				),
			).To(Equal(map[string]string{
				"bucket":    "test2",
				"location":  "US",
				"projectId": "csi-gcs",
			}))
		})
	})
	Describe("MergeMountOptions", func() {
		It("Should Merge", func() {
			Expect(
				MergeMountOptions(
					map[string]string{
						"bucket":   "test",
						"location": "US",
					},
					[]string{"--bucket=test2",
						"--project-id=csi-gcs",
						"--implicit-dirs",
						"--dir-mode=0600",
						"--file-mode=600",
						"--fuse-mount-option=foo,bar",
						"--fuse-mount-option=baz",
					},
				),
			).To(Equal(map[string]string{
				"bucket":           "test2",
				"implicitDirs":     "true",
				"dirMode":          "0600",
				"fileMode":         "0600",
				"location":         "US",
				"fuseMountOptions": "foo,bar,baz",
				"projectId":        "csi-gcs",
			}))
		})
	})
	Describe("ExtraFlags", func() {
		It("Should Merge", func() {
			Expect(
				ExtraFlags(
					map[string]string{
						"bucket":           "test",
						"location":         "US",
						"fuseMountOptions": "foo,bar,baz",
						"implicitDirs":     "true",
						"dirMode":          "0600",
					},
				),
			).To(Equal([]string{"foo", "bar", "baz", "dir_mode=0600", "implicit_dirs"}))
		})
	})
})
