package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

const (
	htpasswdFile = "user.htpasswd"
)

var rootCmd = &cobra.Command{
	Use:   "htpasswd-go",
	Short: "简易htpasswd文件管理CLI工具",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有用户",
	Run: func(cmd *cobra.Command, args []string) {
		users, err := readHtpasswdFile()
		if err != nil {
			fmt.Printf("文件读取失败: %v\n", err)
			return
		}

		for i, user := range users {
			parts := strings.SplitN(user, ":", 2)
			fmt.Printf("%d. %s\n", i+1, parts[0])
		}
	},
}

var addCmd = &cobra.Command{
	Use:   "add [username] [password]",
	Short: "新增一个用户到user.htpasswd文件",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		password := args[1]

		users, err := readHtpasswdFile()
		if err != nil {
			fmt.Printf("文件读取失败: %v\n", err)
			return
		}

		for _, user := range users {
			parts := strings.SplitN(user, ":", 2)
			if parts[0] == username {
				fmt.Printf("用户 %s 已存在, 可以使用修改密码功能\n", username)
				return
			}
		}

		// 生成密码hash
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("密码哈希生成失败: %v\n", err)
			return
		}

		// 组装用户信息
		entry := fmt.Sprintf("%s:%s\n", username, string(hash))

		// 追加用户信息到文件
		f, err := os.OpenFile(htpasswdFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("文件打开失败: %v\n", err)
			return
		}
		defer f.Close()

		if _, err := f.WriteString(entry); err != nil {
			fmt.Printf("文件写入失败: %v\n", err)
			return
		}

		fmt.Printf("用户 %s 已被添加\n", username)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [username]",
	Short: "从user.htpasswd文件删除一个用户",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		users, err := readHtpasswdFile()
		if err != nil {
			fmt.Printf("文件读取失败: %v\n", err)
			return
		}

		// 判断并删除用户
		var newContent []string
		userFound := false
		for _, user := range users {
			if !strings.HasPrefix(user, username+":") {
				newContent = append(newContent, user)
			} else {
				userFound = true
			}
		}

		if !userFound {
			fmt.Printf("用户 %s 不存在\n", username)
			return
		}

		// 回写文件
		err = os.WriteFile(htpasswdFile, []byte(strings.Join(newContent, "\n")+"\n"), 0644)
		if err != nil {
			fmt.Printf("文件写入失败: %v\n", err)
			return
		}

		fmt.Printf("用户 %s 已被删除\n", username)
	},
}

var editPassCmd = &cobra.Command{
	Use:   "editpass [username] [newpassword]",
	Short: "编辑用户密码",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		newPassword := args[1]

		users, err := readHtpasswdFile()
		if err != nil {
			fmt.Printf("文件读取失败: %v\n", err)
			return
		}

		// 生成密码hash
		hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("密码哈希生成失败: %v\n", err)
			return
		}

		// 更新用户密码
		var newContent []string
		userFound := false
		for _, user := range users {
			if strings.HasPrefix(user, username+":") {
				newContent = append(newContent, fmt.Sprintf("%s:%s", username, string(hash)))
				userFound = true
			} else {
				newContent = append(newContent, user)
			}
		}

		if !userFound {
			fmt.Printf("用户 %s 不存在\n", username)
			return
		}

		// 回写
		err = os.WriteFile(htpasswdFile, []byte(strings.Join(newContent, "\n")+"\n"), 0644)
		if err != nil {
			fmt.Printf("文件写入失败: %v\n", err)
			return
		}

		fmt.Printf("用户 %s 密码已更新\n", username)
	},
}

func readHtpasswdFile() ([]string, error) {
	var users []string

	file, err := os.Open(htpasswdFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			users = append(users, line)
		}
	}

	return users, scanner.Err()
}

func main() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(editPassCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
