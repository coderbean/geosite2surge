package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const prefix = "DOMAIN-SUFFIX, "
const gitRepo = "https://github.com/v2fly/domain-list-community.git" // 替换为你的Git仓库地址

func main() {
	// 定义文件夹路径
	gitDir := "domain-list-community-tmp"
	dateDir := filepath.Join(gitDir, "data")
	surgeRulesDir := "surge-rules"

	_ = os.RemoveAll(gitDir) // 清理git目录
	//_ = os.RemoveAll(surgeRulesDir)
	// 拉取Git仓库
	err := cloneGitRepo(gitDir, gitRepo)
	if err != nil {
		fmt.Println("Error cloning Git repository:", err)
		return
	}
	defer os.RemoveAll(gitDir) // 在程序结束时删除data文件夹

	// 创建surge-rules文件夹
	err = os.MkdirAll(surgeRulesDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating surge-rules directory:", err)
		return
	}

	// 读取data文件夹下的所有文件
	files, err := ioutil.ReadDir(dateDir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	// 处理每一个文件
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(dateDir, file.Name())
			outputFilePath := filepath.Join(surgeRulesDir, file.Name()+".list")
			processFile(filePath, outputFilePath)
		}
	}
}

func cloneGitRepo(dir, repo string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", repo, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func processFile(inputFilePath, outputFilePath string) {
	// 打开输入文件
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		fmt.Println("Error opening input file:", err)
		return
	}
	defer inputFile.Close()

	// 创建输出文件
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	writer := bufio.NewWriter(outputFile)

	// 读取输入文件的每一行，处理非空行，并写入输出文件
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			line = prefix + line
		}
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			fmt.Println("Error writing to output file:", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input file:", err)
		return
	}

	// 确保所有内容都写入输出文件
	if err := writer.Flush(); err != nil {
		fmt.Println("Error flushing writer:", err)
		return
	}
}
