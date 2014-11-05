package erasure

/*
#cgo CFLAGS: -Wall
#include "types.h"
#include "erasure_code.h"
*/
import "C"

import (
	"fmt"
	"log"
	"unsafe"
)

type Code struct {
	M            int
	K            int
	VectorLength int
	EncodeMatrix []byte
	galoisTables []byte
}

func NewCode(m int, k int, size int) *Code {
	if m <= 0 || k <= 0 || k >= m || size < 0 {
		log.Fatal("Invalid erasure code params")
	}
	if size%k != 0 {
		log.Fatal("Size to encode is not divisable by k and therefore cannot be enocded in vector chunks")
	}

	encodeMatrix := make([]byte, m*k)
	galoisTables := make([]byte, k*(m-k)*32)

	if k > 5 {
		C.gf_gen_cauchy1_matrix((*C.uchar)(&encodeMatrix[0]), C.int(m), C.int(k))
	} else {
		C.gf_gen_rs_matrix((*C.uchar)(&encodeMatrix[0]), C.int(m), C.int(k))
	}

	return &Code{
		M:            m,
		K:            k,
		VectorLength: size / k,
		EncodeMatrix: encodeMatrix,
		galoisTables: galoisTables,
	}
}

// Data buffer to encode must be of the k*size given in the constructor
// The returned encoded buffer is (m-k)*size, since the first k*size of the
// encoded data is just the original data due to the identity matrix
func (c *Code) Encode(data []byte) []byte {
	if len(data) != c.K*c.VectorLength {
		log.Fatal("Data to encode is not the proper size")
	}
	// Since the first k row of the encode matrix is actually the identity matrix
	// we only need to encode the last m-k vectors of the matrix and append
	// them to the original data
	encoded := make([]byte, (c.M-c.K)*(c.VectorLength))
	C.ec_init_tables(C.int(c.K), C.int(c.M-c.K), (*C.uchar)(&c.EncodeMatrix[c.K*c.K]), (*C.uchar)(&c.galoisTables[0]))
	C.ec_encode_data(C.int(c.VectorLength), C.int(c.K), C.int(c.M-c.K), (*C.uchar)(&c.galoisTables[0]), (*C.uchar)(&data[0]), (*C.uchar)(&encoded[0]))

	// return append(data, encoded...)
	return encoded
}

// Data buffer to decode must be of the k*size given in the constructor
// The source error list must contain m-k values, corresponding to the vectors with errors
// The returned decoded data is k*size
func (c *Code) Decode(encoded []byte, srcErrList []int8) []byte {
	if len(encoded) != c.K*c.VectorLength {
		log.Fatal("Data to decode is not the proper size")
	}
	if len(srcErrList) != c.M-c.K {
		log.Fatal("Err list is not the proper size")
	}
	decodeMatrix := make([]byte, c.M*c.K)
	decodeIndex := make([]int32, c.M)
	srcInErr := make([]int8, c.M)
	nErrs := len(srcErrList)
	nSrcErrs := 0
	for _, err := range srcErrList {
		srcInErr[err] = 1
		if err < int8(c.K) {
			nSrcErrs++
		}
	}

	C.gf_gen_decode_matrix((*C.uchar)(&c.EncodeMatrix[0]), (*C.uchar)(&decodeMatrix[0]), (*C.uint)(unsafe.Pointer(&decodeIndex[0])), (*C.uchar)(unsafe.Pointer(&srcErrList[0])), (*C.uchar)(unsafe.Pointer(&srcInErr[0])), C.int(nErrs), C.int(nSrcErrs), C.int(c.K), C.int(c.M))

	C.ec_init_tables(C.int(c.K), C.int(nErrs), (*C.uchar)(&decodeMatrix[0]), (*C.uchar)(&c.galoisTables[0]))

	recovered := []byte{}
	for i := 0; i <= c.K; i++ {
		recovered = append(recovered, encoded[(decodeIndex[i]*int32(c.VectorLength)):(decodeIndex[i]+1)*int32(c.VectorLength)]...)
	}

	data := make([]byte, c.M*c.VectorLength)
	C.ec_encode_data(C.int(c.VectorLength), C.int(c.K), C.int(c.M), (*C.uchar)(&c.galoisTables[0]), (*C.uchar)(&recovered[0]), (*C.uchar)(&data[0]))

	return data[:c.K*c.VectorLength]
}

func Hello() {
	var m int32 = 12
	var k int32 = 8
	var sourceLength int32 = 16

	source := make([]byte, k*sourceLength)
	destination := make([]byte, m*sourceLength)

	for i := range source {
		source[i] = 0x62
		destination[i] = 0x62
	}

	encodeMatrix := make([]byte, m*k)
	// decode_matrix := make([]byte, m*k)
	// invert_matrix := make([]byte, m*k)
	g_tbls := make([]byte, k*(m-k)*32)

	// fmt.Printf("Encode Matrix: %x\n", encode_matrix)

	// Generate encode matrix encode_matrix
	// The matrix generated by gf_gen_cauchy1_matrix
	// is always invertable.
	C.gf_gen_cauchy1_matrix((*C.uchar)(&encodeMatrix[0]), C.int(m), C.int(k))
	// The matrix generated by gf_gen_rs_matrix
	// is not always invertable.
	// C.gf_gen_rs_matrix((*C.uchar)(&encode_matrix[0]), C.int(m), C.int(k))

	// fmt.Printf("Encode Matrix: %x\n", encodeMatrix)

	// fmt.Printf("G Tables: %x\n", g_tbls)
	C.ec_init_tables(C.int(k), C.int(m-k), (*C.uchar)(&encodeMatrix[k*k]), (*C.uchar)(&g_tbls[0]))
	// fmt.Printf("G Tables: %x\n", g_tbls)

	// fmt.Printf("Source: %x\n", source)
	C.ec_encode_data(C.int(sourceLength), C.int(k), C.int(m-k), (*C.uchar)(&g_tbls[0]), (*C.uchar)(&source[0]), (*C.uchar)(&destination[k*sourceLength]))
	//fmt.Printf("Dest: %x\n", destination)

	decodeMatrix := make([]byte, m*k)
	decodeIndex := make([]int32, m)
	srcErrList := make([]int8, m-k)
	srcInErr := make([]int8, m)

	srcErrList[0] = 0
	srcErrList[1] = 2
	srcErrList[2] = 3
	srcErrList[3] = 4
	srcInErr[0] = 1
	srcInErr[2] = 1
	srcInErr[3] = 1
	srcInErr[4] = 1

	nErrs := 4
	nSrcErrs := 4

	C.gf_gen_decode_matrix((*C.uchar)(&encodeMatrix[0]), (*C.uchar)(&decodeMatrix[0]), (*C.uint)(unsafe.Pointer(&decodeIndex[0])), (*C.uchar)(unsafe.Pointer(&srcErrList[0])), (*C.uchar)(unsafe.Pointer(&srcInErr[0])), C.int(nErrs), C.int(nSrcErrs), C.int(k), C.int(m))
	fmt.Printf("Decode Matrix: %x\n", decodeMatrix)
	// fmt.Printf("Decode Index: %x\n", decodeIndex)

	g_tbls = make([]byte, k*(m-k)*32)

	C.ec_init_tables(C.int(k), C.int(nErrs), (*C.uchar)(&decodeMatrix[0]), (*C.uchar)(&g_tbls[0]))
	// fmt.Printf("G Tables: %x\n", g_tbls)
	returned := []byte{}
	var i int32 = 0
	for i = 0; i <= k; i++ {
		returned = append(returned, destination[(decodeIndex[i]*sourceLength):(decodeIndex[i]+1)*sourceLength]...)
	}

	fmt.Printf("Returned: %x\n", returned)

	recovered := make([]byte, m*sourceLength)
	C.ec_encode_data(C.int(sourceLength), C.int(k), C.int(m), (*C.uchar)(&g_tbls[0]), (*C.uchar)(&returned[0]), (*C.uchar)(&recovered[0]))
	// fmt.Printf("Recovered: %x\n", recovered)
}
