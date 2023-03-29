package cache_test

import (
	"reflect"

	. "github.com/go-web-kits/cache"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Entry", func() {
	Describe("Compress", func() {
		It("compresses string", func() {
			Expect(Compress("abc")).To(Equal("string##abc"))
		})

		It("compresses bool", func() {
			Expect(Compress(true)).To(Equal("bool##true"))
		})

		It("compresses number", func() {
			Expect(Compress(1)).To(Equal("int##1"))
			Expect(Compress(1.2)).To(Equal("float64##1.2"))
			Expect(Compress(uint(1))).To(Equal("uint##1"))
			Expect(Compress(int64(1))).To(Equal("int64##1"))
			Expect(Compress(float64(1))).To(Equal("float64##1"))
			Expect(Compress(byte(1))).To(Equal("uint8##1"))
		})

		It("compresses array", func() {
			Expect(Compress([2]byte{1, 2})).To(Equal("[2]uint8##[1,2]"))
			Expect(Compress([1][2]int{{3, 4}})).To(Equal("[1][2]int##[[3,4]]"))
			Expect(Compress([3]string{"a", "b", "c"})).To(Equal("[3]string##[\"a\",\"b\",\"c\"]"))
		})

		It("compresses slice", func() {
			Expect(Compress([]float64{1.1, 2.2})).To(Equal("[]float64##[1.1,2.2]"))
		})

		It("compresses struct", func() {
			Expect(Compress(struct{ Abc string }{Abc: "abc"})).To(Equal("struct { Abc string }##{\"Abc\":\"abc\"}"))
			type MyStruct struct{ Abc string }
			Expect(Compress(MyStruct{Abc: "abc"})).To(Equal("cache_test.MyStruct##{\"Abc\":\"abc\"}"))
			type NestedStruct struct{ MyStruct }
			Expect(Compress(NestedStruct{MyStruct{Abc: "abc"}})).To(Equal("cache_test.NestedStruct##{\"Abc\":\"abc\"}"))
		})

		It("compresses map", func() {
			Expect(Compress(map[string]interface{}{"a": "b", "c": 1, "d": 2.1})).
				To(Equal("map[string]interface {}##{\"a\":\"b\",\"c\":1,\"d\":2.1}"))
			Expect(Compress(map[string]string{"a": "b"})).
				To(Equal("map[string]string##{\"a\":\"b\"}"))
			Expect(Compress(map[int]string{9: "b"})).
				To(Equal("map[int]string##{\"9\":\"b\"}"))
		})
	})

	Describe("UnCompress", func() {
		var InAndOut = func(value interface{}, obj ...interface{}) (interface{}, error) {
			c, _ := Compress(value)
			return UnCompress(c, obj...)
		}

		var CanProcess = func(value interface{}, obj ...interface{}) bool {
			if len(obj) > 0 {
				result, _ := InAndOut(value, obj...)
				return Expect(reflect.Indirect(reflect.ValueOf(result)).Interface()).To(Equal(value))
			} else {
				return Expect(InAndOut(value, obj...)).To(Equal(value))
			}
		}

		It("un-compresses string", func() {
			CanProcess("abc")
		})

		It("un-compresses bool", func() {
			CanProcess(true)
		})

		It("un-compresses number", func() {
			CanProcess(1)
			CanProcess(1.2)
			CanProcess(uint(1))
			CanProcess(int64(1))
			CanProcess(float64(1))
			CanProcess(byte(1))
		})

		Context("un-compresses array & slice", func() {
			It("currently returns []interface{}", func() {
				Expect(InAndOut([2]string{"a", "b"})).To(Equal([]interface{}{"a", "b"}))
				Expect(InAndOut([]string{"a", "b"})).To(Equal([]interface{}{"a", "b"}))

				Expect(InAndOut([1][2]int{{3, 4}})).To(Equal([]interface{}{
					[]interface{}{float64(3), float64(4)},
				}))
			})

			It("currently makes the number to float64", func() {
				Expect(InAndOut([2]byte{1, 2})).To(Equal([]interface{}{float64(1), float64(2)}))
				Expect(InAndOut([]float64{1.3})).To(Equal([]interface{}{float64(1.3)}))
			})

			When("passing object", func() {
				It("unmarshal to the object and returns the object", func() {
					var bs [2]byte
					CanProcess([2]byte{1, 2}, &bs)

					var s [][]string
					CanProcess([][]string{{"a", "b"}}, &s)
				})
			})
		})

		Context("un-compresses struct", func() {
			It("currently returns map[string]interface{}", func() {
				Expect(InAndOut(struct{ Abc string }{Abc: "abc"})).To(Equal(map[string]interface{}{"Abc": "abc"}))
			})

			It("currently makes the number to float64", func() {
				Expect(InAndOut(struct{ Abc int }{Abc: 1})).To(Equal(map[string]interface{}{"Abc": 1.0}))
			})

			When("passing object", func() {
				It("unmarshal to the object and returns the object", func() {
					type MyStruct struct{ Abc int }
					type NestedStruct struct{ MyStruct }
					var obj NestedStruct
					CanProcess(NestedStruct{MyStruct{Abc: 1}}, &obj)
				})
			})
		})

		It("un-compresses map", func() {
			CanProcess(map[string]interface{}{"a": "b", "c": float64(1), "d": 2.1})

			var m map[int]float64
			CanProcess(map[int]float64{123: 4.56}, &m)
		})
	})
})
