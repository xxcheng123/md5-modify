package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var start_path string
var append_content string
var success_count = 0
var fail_count = 0
var file_index = 0

// var max_thread_num = 10

func visit(path string, info os.FileInfo, err error) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("处理[%s]时发生错误\n", info.Name())
		} else if !info.IsDir() {
			fail_count--
			success_count++
		}
	}()
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	if !info.IsDir() {
		fail_count++

		md5, _ := calculateMD5(path)
		appendFileToChar(path, []byte(append_content))
		modifed_md5, _ := calculateMD5(path)
		file_index++
		current_index := file_index
		fmt.Printf("%d.【%s】 修改md5成功：[%s]=>[%s]\n", current_index, info.Name(), md5, modifed_md5)
	}

	return nil
}
func calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
func appendFileToChar(filepath string, bytes []byte) bool {
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.Write(bytes)
	if err != nil {
		panic(err)
	} else {
		return true
	}
}

func init() {
	flag.StringVar(&start_path, "p", "", "遍历修改起始路径（默认为当前路径）")
	flag.StringVar(&append_content, "s", "#123", "在末尾添加的字符（默认为#123）")
	flag.Parse()
	if start_path == "" {
		dir, _ := os.Getwd()
		start_path = dir
	} else {
		path, err := filepath.Abs(start_path)
		if err != nil {
			panic("提供的路径错误")
		}
		start_path = path
	}
	fmt.Println("起始路径：" + start_path)
	var is_next string

	for {
		fmt.Print("是否确认执行？[Y/n]:")
		fmt.Scanln(&is_next)
		is_next = strings.ToLower(is_next)
		if is_next == "n" || is_next == "no" {
			fmt.Println("bye~")
			os.Exit(0)
		} else if is_next == "y" || is_next == "yes" {
			break
		}
		fmt.Println("指令错误，请重新输入~")
	}

}

func outEndInfo() {
	fmt.Printf("成功处理[%d]个文件,失败处理[%d]个文件\n", success_count, fail_count)
}
func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("程序执行错误,%#v\n", err)
		}
	}()
	defer outEndInfo()
	defer fmt.Print("程序执行完毕，")
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)

		<-c
		fmt.Print("用户手动退出，")
		outEndInfo()
		os.Exit(0)
	}()
	fmt.Println("开始运行~")
	startTime := time.Now()
	// filepath.Walk(start_path, visit)
	var wg sync.WaitGroup
	filepath.Walk(start_path, func(path string, info fs.FileInfo, err error) error {
		wg.Add(1)
		go func() {
			defer wg.Done()
			visit(path, info, err)
		}()
		return nil
	})
	wg.Wait()
	elapsedTime := time.Since(startTime) / time.Millisecond
	fmt.Printf("执行耗时：%dms\n", elapsedTime)
}
