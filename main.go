package main

import (
	"flag"
	"fmt"
	. "io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

func readAll(path string) []string {
	var all_file []string
	finfo, _ := ReadDir(path)
	for _, x := range finfo {
		real_path := path + "/" + x.Name()
		if !x.IsDir() {
			all_file = append(all_file, real_path)
		}
	}
	return all_file
}

var (
	CHARS string
	mode  int32
)

func init() {
	flag.StringVar(&CHARS, "chars", "", "需要生成的文字,支持多个文字生成")
}

func main() {
	flag.Parse()

	code := 0
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("error:", err)
			fmt.Printf("生成失败：%s", string(code))
		}
	}()

	mode = CONFIG.GetInt32("", "mode")
	if CHARS == "" {
		// 生成汉字
		files := readAll(CONFIG.GetString("", "data_dir"))
		for _, file := range files {
			_, fileName := filepath.Split(file)
			name := strings.Split(fileName, ".")[0]

			if mode == 1 {
				fmt.Printf("生成图片：%s\n", name)
				genImage(name)
			} else {
				code, _ = strconv.Atoi(name)
				fmt.Printf("生成文字：%s\n", string(code))
				genPictures(string(code))
			}
		}
	} else {
		if mode == 1 {
			fmt.Printf("生成图片：%s\n", CHARS)
			genImage(CHARS)
		} else {
			chars := strings.Split(CHARS, "")
			for _, c := range chars {
				code = int(getCode(c))
				fmt.Printf("生成文字：%s\n", c)
				genPictures(c)
			}
		}
	}
}
