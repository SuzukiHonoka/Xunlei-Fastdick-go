package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func GetStringMd5Hex(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func GetStringSha1Hex(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func getMacAddr() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	return interfaces[0].HardwareAddr.String(), nil
}

func ForceMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func ForceClose(io io.Closer) {
	_ = io.Close()
}

func Question(q string) string {
	fmt.Print(q)
	var a string
	_, _ = fmt.Scanf("%s", &a)
	return a
}

func JsonToMap(b []byte) map[string]interface{} {
	var infos map[string]interface{}
	_ = json.Unmarshal(b, &infos)
	return infos
}
