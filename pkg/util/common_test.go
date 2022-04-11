package util_test

import (
	"io/ioutil"
	"os"

	. "github.com/ofek/csi-gcs/pkg/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Common", func() {
	Describe("GetKey", func() {
		It("should error when neither 'key' nor 'key.json' are present", func() {
			sec := map[string]string{}

			_, err := GetKey(sec, "")
			Expect(err).ShouldNot(Equal("Secret has no keys named 'key' or 'key.json'"))
		})

		It("should use 'key' first", func() {
			sec := map[string]string{
				"key":      "Content of key",
				"key.json": "Content of key.json",
			}

			dir, err := ioutil.TempDir("", "test")
			Expect(err).ShouldNot(HaveOccurred())
			defer os.RemoveAll(dir)

			s, err := GetKey(sec, dir)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(s).To(BeARegularFile())

			f, err := os.Open(s)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ioutil.ReadAll(f)).To(BeEquivalentTo("Content of key"))
		})

		It("should fallback to 'key.json'", func() {
			sec := map[string]string{
				"key.json": "Content of key.json",
			}

			dir, err := ioutil.TempDir("", "test")
			Expect(err).ShouldNot(HaveOccurred())
			defer os.RemoveAll(dir)

			s, err := GetKey(sec, dir)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(s).To(BeARegularFile())

			f, err := os.Open(s)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ioutil.ReadAll(f)).To(BeEquivalentTo("Content of key.json"))
		})
	})
})
