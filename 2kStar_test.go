/*
 * @Author: 2Kil
 * @Date: 2024-04-19 10:54:20
 * @LastEditors: 2Kil
 * @LastEditTime: 2024-05-28 17:49:00
 * @Description:star
 */

package star

import (
	"log"
	"testing"
)

func TestRandAtomicRadius(t *testing.T) {
	type args struct {
		min int
		max int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
		{"t1", args{10, 14}, 100},
		{"t1", args{10, 14}, 100},
		{"t1", args{10, 14}, 100},
		{"t1", args{10, 14}, 100},
		{"t1", args{10, 14}, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandAtomicRadius(tt.args.min, tt.args.max)
			log.Println(got)
		})
	}
}
