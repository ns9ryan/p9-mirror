// 演示 oa.98ent.com/p9/common/modules/fieldcrypto：加密入库、按规范化值检索。
//
//	go run ./modules/fieldcrypto/example
package main

import (
	"crypto/rand"
	"fmt"
	"os"

	"oa.98ent.com/p9/common/modules/fieldcrypto"
)

func main() {
	encKey := make([]byte, 32)
	macKey := make([]byte, 32)
	if _, err := rand.Read(encKey); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err := rand.Read(macKey); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	c, err := fieldcrypto.New(encKey, macKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	phonePlain := "60-123 456 789"
	emailPlain := "  Foo@Bar.COM "

	phoneCT, err := c.Seal(phonePlain)
	must(err)
	emailCT, err := c.Seal(emailPlain)
	must(err)

	phoneHash, err := c.LookupHash(fieldcrypto.NormalizePhone(phonePlain))
	must(err)
	emailHash, err := c.LookupHash(fieldcrypto.NormalizeEmail(emailPlain))
	must(err)

	fmt.Println("== 写入（密文不同、hash 可查询）")
	fmt.Println("phone ciphertext:", phoneCT)
	fmt.Println("phone lookup hash:", phoneHash)
	fmt.Println("email ciphertext:", emailCT)
	fmt.Println("email lookup hash:", emailHash)

	gotPhone, err := c.Open(phoneCT)
	must(err)
	gotEmail, err := c.Open(emailCT)
	must(err)
	fmt.Println("== 解密")
	fmt.Println("phone:", gotPhone)
	fmt.Println("email:", gotEmail)

	query := fieldcrypto.NormalizePhone("+60-123 456 789")
	qHash, err := c.LookupHash(query)
	must(err)
	fmt.Println("== 按规范化手机号检索")
	fmt.Println("query:", query)
	fmt.Println("match:", qHash == phoneHash)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
