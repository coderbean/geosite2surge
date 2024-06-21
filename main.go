package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const domainSuffixPrefix = "DOMAIN-SUFFIX, "
const domainPrefix = "DOMAIN, "
const urlRegexPrefix = "URL-REGEX, "
const gitRepo = "https://github.com/v2fly/domain-list-community.git" // 替换为你的Git仓库地址
var (
	processFlag = map[string]bool{}
	// 定义文件夹路径
	gitDir        = "domain-list-community-tmp"
	dateDir       = filepath.Join(gitDir, "data")
	surgeRulesDir = "surge-rules"
)

func main() {

	// 拉取Git仓库
	_ = os.RemoveAll(gitDir) // 清理git目录
	err := cloneGitRepo(gitDir, gitRepo)
	if err != nil {
		fmt.Println("Error cloning Git repository:", err)
		return
	}
	defer os.RemoveAll(gitDir) // 在程序结束时删除data文件夹

	// 创建surge-rules文件夹
	createErr := os.MkdirAll(surgeRulesDir, os.ModePerm)
	if createErr != nil {
		fmt.Println("Error creating surge-rules directory:", createErr)
		return
	}

	// 读取data文件夹下的所有文件
	files, err := os.ReadDir(dateDir)
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
	if processFlag[inputFilePath] {
		return
	}
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
		line := removeAtAndAfter(scanner.Text())
		trimLine := strings.TrimSpace(line)
		if trimLine != "" {
			if strings.HasPrefix(trimLine, "#") {
				// 保持注释不动
			} else if strings.HasPrefix(trimLine, "full:") {
				line = strings.ReplaceAll(line, "full:", domainPrefix)
			} else if strings.HasPrefix(trimLine, "regexp:") {
				line = strings.ReplaceAll(line, "regexp:", urlRegexPrefix)
			} else if strings.HasPrefix(trimLine, "include:") {
				// 递归处理
				subInputFilePath, subOutPutFilePath := genInputAndOutputFileName(strings.Split(trimLine[len("include:"):], " ")[0])
				processFile(subInputFilePath, subOutPutFilePath)
				// 处理完成后将整个文件替换该行
				lineBytes, _ := os.ReadFile(subOutPutFilePath)
				line = string(lineBytes)
			} else {
				line = domainSuffixPrefix + line
			}
		}
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			fmt.Println("Error writing to output file:", err)
			return
		}
		processFlag[inputFilePath] = true
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

// removeAtAndAfter 删除字符串中 '@' 及其后面的所有字符
func removeAtAndAfter(input string) string {
	input = strings.TrimSpace(input)
	if idx := strings.Index(input, " @"); idx != -1 {
		return input[:idx]
	}
	return input
}

func genInputAndOutputFileName(key string) (string, string) {
	inputFilePath := filepath.Join(dateDir, key)
	outputFilePath := filepath.Join(surgeRulesDir, key+".list")
	return inputFilePath, outputFilePath
}
