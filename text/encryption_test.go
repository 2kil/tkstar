/*
 * @Author: 2Kil
 * @Date: 2024-05-27 23:07:41
 * @LastEditors: 2Kil
 * @LastEditTime: 2024-05-28 00:08:27
 * @Description: 文本加密解密
 */

package star

import (
	"log"
	"testing"
)

func TestEorEncode(t *testing.T) {
	type args struct {
		text string
		key  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"1", args{"1234567890", "1234567890"}, "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EorEncode(tt.args.text, tt.args.key)
			log.Println(got)
		})
	}
}

func TestEorDecode(t *testing.T) {
	type args struct {
		ciphertext string
		key        string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"1", args{"fGZ6TntyYgp3SipoAAxwcA==", "1234567890"}, "1234567890", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EorDecode(tt.args.ciphertext, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("EorDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			log.Println(got)
		})
	}
}
