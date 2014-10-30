package erasure

/*
#cgo CFLAGS: -Wall
#include "types.h"
#include "erasure_code.h"
*/
import "C"

import (
	"fmt"
)

func Hello() {
	fmt.Println("Hello")

	m := 12
	k := 8

	sources := make([][]byte, 127)
	encode_matrix := make([]byte, m*k)
	// decode_matrix := make([]byte, m*k)
	// invert_matrix := make([]byte, m*k)
	g_tbls := make([]byte, k*len(sources)*32)

	// fmt.Printf("Encode Matrix: %x\n", encode_matrix)
	C.gf_gen_rs_matrix((*C.uchar)(&encode_matrix[0]), C.int(m), C.int(k))
	// fmt.Printf("Encode Matrix: %x\n", encode_matrix)

	// fmt.Printf("G Tables: %x\n", g_tbls)
	C.ec_init_tables(C.int(k), C.int(m-k), (*C.uchar)(&encode_matrix[k*k]), (*C.uchar)(&g_tbls[0]))
	// fmt.Printf("G Tables: %x\n", g_tbls)
}
