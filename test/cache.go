package cache_test

import (
	"time"

	. "github.com/go-web-kits/cache"
	"github.com/go-web-kits/dbx"
	. "github.com/go-web-kits/testx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("Cache", func() {
	var (
		p *MonkeyPatches
	)

	AfterEach(func() {
		p.Check()
	})

	Describe("DeleteMatched", func() {
		It("does successfully", func() {
			Expect(Set("key1", "test")).To(Succeed())
			Expect(DeleteMatched("key*")).To(Succeed())
			Expect(DeleteMatched("fetch*")).To(Succeed())
			_, err := Get("key1")
			Expect(err).To(HaveOccurred())
		})

		When("no key matched", func() {
			It("does nothing", func() {
				Expect(DeleteMatched("xqwe*")).To(Succeed())
			})
		})
	})

	Describe("Set", func() {
		It("does successfully", func() {
			Expect(Set("key1", 1.2)).To(Succeed())
			Expect(Set("key2", "string")).To(Succeed())
			Expect(Set("key3", struct{ Abc string }{Abc: "abc"})).To(Succeed())
		})

		When("re-calling it", func() {
			It("recovers", func() {
				Expect(Set("key4", 1)).To(Succeed())
				Expect(Set("key4", map[string]interface{}{"hello": "world"})).To(Succeed())
				Expect(Get("key4")).To(Equal(map[string]interface{}{"hello": "world"}))
			})
		})
	})

	Describe("Get", func() {
		It("does successfully", func() {
			Expect(Get("key1")).To(Equal(1.2))
			Expect(Get("key2")).To(Equal("string"))
			Expect(Get("key3")).To(Equal(map[string]interface{}{"Abc": "abc"}))
		})

		It("does successfully with opt.To", func() {
			var m map[string]interface{}
			Expect(Get("key3", Opt{To: &m})).To(Equal(&m))
			Expect(m).To(Equal(map[string]interface{}{"Abc": "abc"}))
			var obj struct{ Abc string }
			Expect(Get("key3", Opt{To: &obj})).To(Equal(&obj))
			Expect(obj).To(Equal(struct{ Abc string }{Abc: "abc"}))
		})

		When("the key not found", func() {
			It("returns error", func() {
				_, err := Get("xxxyz")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Fetch", func() {
		var (
			isExpectedNotToCallSet = func() {
				p = IsExpectedToCall(Set).AndReturn(nil).NotOnce()
			}
		)

		When("the key exists", func() {
			It("returns the value directly", func() {
				isExpectedNotToCallSet()
				Expect(Fetch("key1")).To(Equal(1.2))
			})

			It("unmarshal to the passed object", func() {
				isExpectedNotToCallSet()
				var obj struct{ Abc string }
				Expect(Fetch("key3", Opt{To: &obj})).To(Equal(&obj))
				Expect(obj).To(Equal(struct{ Abc string }{Abc: "abc"}))
			})
		})

		When("the key does not exist", func() {
			It("returns nil (no error) if not Force", func() {
				isExpectedNotToCallSet()
				Expect(Fetch("fetch1")).To(BeNil())
			})

			It("returns not found error if Force", func() {
				isExpectedNotToCallSet()
				_, err := Fetch("fetch2", Opt{Force: true})
				Expect(err).To(HaveOccurred())
			})

			Context("but passing Default value", func() {
				Context("which type is not func", func() {
					It("Sets the value", func() {
						Expect(Fetch("fetch3", Opt{Default: "true"})).To(Equal("true"))
						Expect(Get("fetch3")).To(Equal("true"))
					})
				})

				Context("which type is `func() interface{}`", func() {
					It("executes the function and Sets the result", func() {
						Expect(Fetch("fetch4", Opt{Default: func() interface{} {
							return 1.2
						}})).To(Equal(1.2))
						Expect(Get("fetch4")).To(Equal(1.2))

						Expect(Fetch("fetch5", Opt{Default: func() interface{} {
							return dbx.Result{Data: []byte("hello")}
						}})).To(Equal([]uint8("hello")))
					})

					It("executes the function and but gets error, do nothing", func() {
						isExpectedNotToCallSet()

						_, err1 := Fetch("fetch6", Opt{Default: func() interface{} {
							return errors.New("")
						}})
						Expect(err1).To(HaveOccurred())

						_, err2 := Fetch("fetch7", Opt{Default: func() interface{} {
							return dbx.Result{Err: errors.New("")}
						}})
						Expect(err2).To(HaveOccurred())
					})
				})
			})
		})
	})

	Describe("Delete", func() {
		It("does successfully", func() {
			Expect(Set("key1", "test")).To(Succeed())
			Expect(Delete("key1")).To(Succeed())
			_, err := Get("key1")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Increase", func() {
		It("does successfully", func() {
			Expect(Delete("ic1")).To(Succeed())
			Expect(Increase("ic1")).To(Succeed())
			Expect(Get("ic1")).To(Equal("1"))

			Expect(Increase("ic1", 10)).To(Succeed())
			Expect(Get("ic1")).To(Equal("11"))
		})
	})

	Describe("Decrease", func() {
		It("does successfully", func() {
			Expect(Delete("dc1")).To(Succeed())
			Expect(Decrease("dc1")).To(Succeed())
			Expect(Get("dc1")).To(Equal("-1"))

			Expect(Decrease("dc1", 10)).To(Succeed())
			Expect(Get("dc1")).To(Equal("-11"))
		})
	})

	Describe("DistributedLock", func() {
		Measure("it is a spin lock", func(b Benchmarker) {
			Expect(Delete("lock")).To(Succeed())

			lock, err := GetLock("lock", 500*time.Millisecond)
			Expect(err).NotTo(HaveOccurred())

			runtime := b.Time("runtime", func() {
				Expect(Set("lock", "string", Opt{UnderLocking: true, FailIfLocked: true})).NotTo(Succeed())
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.1))
			Expect(lock.Release()).To(Succeed())
		}, 1)

		Measure("it should fail when the lock is not released", func(b Benchmarker) {
			Expect(Delete("lock")).To(Succeed())

			lock, err := GetLock("lock", 500*time.Millisecond)
			Expect(err).NotTo(HaveOccurred())

			runtime := b.Time("runtime", func() {
				Expect(Set("lock", "string", Opt{UnderLocking: true})).To(Succeed())
			})

			Expect(runtime.Seconds()).Should(BeNumerically(">", 0.4))
			Expect(lock.Release()).NotTo(Succeed())
		}, 1)
	})
})
