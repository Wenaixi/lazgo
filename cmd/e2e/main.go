package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/lazgo/lazgo/pkg/auth"
	"github.com/lazgo/lazgo/pkg/client"
	"github.com/lazgo/lazgo/pkg/files"
	"github.com/lazgo/lazgo/pkg/folders"
	"github.com/lazgo/lazgo/pkg/recycle"
	"github.com/lazgo/lazgo/pkg/share"
)

// 命令行标志
var (
	// 测试选项
	testLogin   = flag.Bool("login", false, "通过账密登录测试")
	testBatch  = flag.Bool("batch", false, "执行批量操作测试")
	testClear  = flag.Bool("clear", false, "执行清空回收站测试")
	showHelp  = flag.Bool("h", false, "显示帮助")
)

// 凭证路径
var credPath = "data/lanzou_credentials.json"
var cookiePath = "data/lanzou_cookies.json"

// ==================== 核心函数 ====================

// getCookiePath 返回 cookie 文件路径
func getCookiePath() string {
	if path := os.Getenv("LAZGO_COOKIE_FILE"); path != "" {
		return path
	}
	searchPaths := []string{
		"../data/lanzou_cookies.json",
		"data/lanzou_cookies.json",
		"../../data/lanzou_cookies.json",
	}
	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return searchPaths[0]
}

// getCredPath 返回凭证文件路径
func getCredPath() string {
	searchPaths := []string{
		"../data/lanzou_credentials.json",
		"data/lanzou_credentials.json",
		"../../data/lanzou_credentials.json",
	}
	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return searchPaths[0]
}

// ensureLoggedIn 从 cookie 登录
func ensureLoggedIn() (*client.Client, error) {
	data, err := ioutil.ReadFile(getCookiePath())
	if err != nil {
		return nil, err
	}
	var cookies map[string]string
	if err := json.Unmarshal(data, &cookies); err != nil {
		return nil, err
	}
	if len(cookies) == 0 {
		return nil, fmt.Errorf("no cookies")
	}
	c := client.New(cookies)
	if c.UID() == "" {
		return nil, fmt.Errorf("invalid cookies")
	}
	return c, nil
}

// loginWithCred 从账密登录
func loginWithCred() (*client.Client, error) {
	credFile := getCredPath()
	data, err := ioutil.ReadFile(credFile)
	if err != nil {
		return nil, fmt.Errorf("读取凭证文件失败: %v", err)
	}
	var cred struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil, fmt.Errorf("解析凭证失败: %v", err)
	}
	if cred.Username == "" || cred.Password == "" {
		return nil, fmt.Errorf("凭证文件缺少 username 或 password")
	}

	fmt.Printf("  📝 使用账密登录: %s\n", cred.Username)
	cookies, err := auth.Login(cred.Username, cred.Password)
	if err != nil {
		return nil, fmt.Errorf("登录失败: %v", err)
	}

	// 保存 cookie
	cookieData, _ := json.MarshalIndent(cookies, "", "  ")
	if err := ioutil.WriteFile(getCookiePath(), cookieData, 0644); err != nil {
		fmt.Printf("  ⚠️  保存 cookie 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ Cookie 已保存\n")
	}

	return client.New(cookies), nil
}

// createTempFile 创建临时文件
func createTempFile(content []byte) (string, error) {
	tmpDir := os.TempDir()
	name := fmt.Sprintf("lazgo_e2e_%d.txt", time.Now().UnixNano())
	tmpFile := filepath.Join(tmpDir, name)
	if err := ioutil.WriteFile(tmpFile, content, 0644); err != nil {
		return "", err
	}
	return tmpFile, nil
}

// ==================== 主流程 ====================

func main() {
	flag.Parse()
	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	fmt.Println("========================================")
	fmt.Println("  lazgo E2E 完整测试")
	fmt.Println("========================================")
	fmt.Println()

	// 根据标志决定登录方式
	var c *client.Client
	var err error

	if *testLogin {
		// 通过账密登录测试
		c, err = loginWithCred()
		if err != nil {
			fmt.Printf("  ❌ 登录失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✅ 登录成功: UID=%s\n\n", c.UID())
	} else {
		// 使用现有 cookie
		c, err = ensureLoggedIn()
		if err != nil {
			fmt.Printf("  ❌ 登录失败: %v\n", err)
			fmt.Printf("  💡 使用 -login 标志可用账密登录\n")
			os.Exit(1)
		}
		fmt.Printf("  ✅ 已登录: UID=%s\n\n", c.UID())
	}

	// 保存原始状态，用于异常恢复
	originalItems, _ := recycle.ListRecycle(c, 1)
	originalCount := len(originalItems)

	// ========== 开始测试 ==========

	// 第一阶段：基础操作（必定执行）
	folderID, fileID, fileID2, subFolderID := runBasicTests(c)
	safeDelay("基础操作完成后")

	// 第二阶段：分享操作（必定执行）
	runShareTests(c, fileID)
	safeDelay("分享操作完成后")

	// 第三阶段：回收站操作（必定执行）
	runRecycleTests(c, fileID, fileID2, subFolderID, originalCount)
	safeDelay("回收站操作完成后")

	// 第四阶段：批量操作（可选）
	if *testBatch {
		runBatchTests(c, folderID)
		safeDelay("批量操作完成后")
	}

	// 第五阶段：危险操作（可选，需明确标志）
	if *testClear {
		runDangerousTests(c)
	}

	// 最终清理
	cleanup(c, folderID, subFolderID)

	fmt.Println("========================================")
	fmt.Println("  ✅ E2E 测试全部通过!")
	fmt.Println("========================================")
	printStats()
}

// ==================== 测试阶段 ====================

// runBasicTests 执行基础测试
func runBasicTests(c *client.Client) (folderID, fileID, fileID2, subFolderID int) {
	fmt.Println("--- A. 文件夹操作 ---")

	// A1: 创建测试文件夹
	fmt.Println("[A1] 创建测试文件夹...")
	folderName := fmt.Sprintf("e2e_test_%d", time.Now().Unix())
	folderInfo, err := folders.CreateFolder(c, folderName, 0, "E2E测试")
	if err != nil {
		panic(fmt.Errorf("创建失败: %v", err))
	}
	folderID = parseFileID(folderInfo.ID)
	fmt.Printf("  ✅ folder_id=%d, name=%s\n", folderID, folderName)

	// A2: 获取文件夹信息
	fmt.Println("[A2] 获取文件夹信息...")
	info, err := folders.GetFolderInfo(c, folderID)
	if err != nil {
		panic(fmt.Errorf("获取失败: %v", err))
	}
	fmt.Printf("  ✅ 文件夹名: %s\n", info.Name)

	// A3: 创建子文件夹
	fmt.Println("[A3] 创建子文件夹...")
	subInfo, err := folders.CreateFolder(c, "子文件夹", folderID, "")
	if err != nil {
		panic(fmt.Errorf("创建子文件夹失败: %v", err))
	}
	subFolderID = parseFileID(subInfo.ID)
	fmt.Printf("  ✅ sub_folder_id=%d\n", subFolderID)

	fmt.Println("\n--- B. 文件操作 ---")

	// B1: 上传测试文件
	fmt.Println("[B1] 上传文件...")
	tmpFile, err := createTempFile([]byte("lazgo e2e test " + time.Now().Format(time.RFC3339)))
	if err != nil {
		panic(fmt.Errorf("创建临时文件失败: %v", err))
	}
	defer os.Remove(tmpFile)

	info2, err := files.Upload(c, tmpFile, folderID)
	if err != nil {
		panic(fmt.Errorf("上传失败: %v", err))
	}
	fileID = parseFileID(info2.ID)
	fmt.Printf("  ✅ file_id=%d, size=%s\n", fileID, info2.Size)

	// B2: 上传第二个文件（用于后续测试）
	fmt.Println("[B2] 上传第二个文件...")
	tmpFile2, _ := createTempFile([]byte("second file"))
	defer os.Remove(tmpFile2)
	info3, _ := files.Upload(c, tmpFile2, folderID)
	fileID2 = parseFileID(info3.ID)
	fmt.Printf("  ✅ file_id=%d\n", fileID2)

	// B3: 列文件
	fmt.Println("[B3] 列文件...")
	list, err := files.ListFiles(c, folderID, 1)
	if err != nil {
		panic(fmt.Errorf("列出失败: %v", err))
	}
	fmt.Printf("  ✅ 文件夹内 %d 项\n", len(list))

	// 根目录
	listRoot, _ := files.ListFiles(c, -1, 1)
	fmt.Printf("  ✅ 根目录 %d 项\n", len(listRoot))

	return folderID, fileID, fileID2, subFolderID
}

// runShareTests 执行分享测试
func runShareTests(c *client.Client, fileID int) {
	fmt.Println("\n--- C. 分享操作 ---")

	// C1: 获取分享链接
	fmt.Println("[C1] 获取分享链接...")
	link, err := share.GetShareLink(c, fileID)
	if err != nil {
		panic(fmt.Errorf("获取分享链接失败: %v", err))
	}
	fmt.Printf("  ✅ %s\n", link.URL)

	// C2: 直链转换
	fmt.Println("[C2] 直链转换...")
	direct, err := share.ShareToDirect(link.URL, "")
	if err != nil {
		panic(fmt.Errorf("直链转换失败: %v", err))
	}
	shortURL := direct.URL
	if len(shortURL) > 50 {
		shortURL = shortURL[:50] + "..."
	}
	fmt.Printf("  ✅ %s\n", shortURL)
}

// runRecycleTests 执行回收站测试
func runRecycleTests(c *client.Client, fileID, fileID2, subFolderID, originalCount int) {
	fmt.Println("\n--- D. 回收站操作 ---")

	// D1: 软删除文件
	fmt.Println("[D1] 软删除文件...")
	if err := files.DeleteFile(c, fileID); err != nil {
		panic(fmt.Errorf("删除失败: %v", err))
	}
	fmt.Printf("  ✅ file_id=%d → 回收站\n", fileID)

	// D2: 软删除第二个文件
	fmt.Println("[D2] 软删除第二个文件...")
	if err := files.DeleteFile(c, fileID2); err != nil {
		panic(fmt.Errorf("删除失败: %v", err))
	}
	fmt.Printf("  ✅ file_id=%d → 回收站\n", fileID2)

	// D3: 软删除子文件夹
	fmt.Println("[D3] 软删除子文件夹...")
	if err := folders.DeleteFolder(c, subFolderID); err != nil {
		panic(fmt.Errorf("删除失败: %v", err))
	}
	fmt.Printf("  ✅ folder_id=%d → 回收站\n", subFolderID)

	// D4: 列出回收站
	fmt.Println("[D4] 列出回收站...")
	items, err := recycle.ListRecycle(c, 1)
	if err != nil {
		panic(fmt.Errorf("列出失败: %v", err))
	}
	newCount := len(items) - originalCount
	fmt.Printf("  ✅ 新增 %d 项，共 %d 项\n", newCount, len(items))

	// 验证在回收站
	var fileFound, folderFound bool
	for _, item := range items {
		if !item.IsFolder && parseFileID(item.ID) == fileID {
			fileFound = true
		}
		if item.IsFolder && parseFileID(item.ID) == subFolderID {
			folderFound = true
		}
	}
	fmt.Printf("  ✅ 文件在回收站: %v, 文件夹在回收站: %v\n", fileFound, folderFound)

	// D5: 恢复文件
	fmt.Println("[D5] 恢复文件...")
	if err := recycle.RestoreFile(c, fileID); err != nil {
		panic(fmt.Errorf("恢复失败: %v", err))
	}
	fmt.Printf("  ✅ file_id=%d 已恢复\n", fileID)

	// D6: 恢复���件���
	fmt.Println("[D6] 恢复文件夹...")
	if err := recycle.RestoreFolder(c, subFolderID); err != nil {
		panic(fmt.Errorf("恢复失败: %v", err))
	}
	fmt.Printf("  ✅ folder_id=%d 已恢复\n", subFolderID)

	// D7: 再次删除（准备永久删除）
	fmt.Println("[D7] 再次删除...")
	if err := files.DeleteFile(c, fileID); err != nil {
		panic(fmt.Errorf("删除失败: %v", err))
	}
	if err := folders.DeleteFolder(c, subFolderID); err != nil {
		panic(fmt.Errorf("删除失败: %v", err))
	}
	fmt.Printf("  ✅ 已删除\n")

	// D8: 永久删除文件
	fmt.Println("[D8] 永久删除文件...")
	if err := recycle.DeleteRecycleFile(c, fileID); err != nil {
		panic(fmt.Errorf("永久删除失败: %v", err))
	}
	fmt.Printf("  ✅ file_id=%d 永久删除\n", fileID)

	// D9: 永久删除文件夹
	fmt.Println("[D9] 永久删除文件夹...")
	if err := recycle.DeleteRecycleFolder(c, subFolderID); err != nil {
		panic(fmt.Errorf("永久删除失败: %v", err))
	}
	fmt.Printf("  ✅ folder_id=%d 永久删除\n", subFolderID)
}

// runBatchTests 执行批量操作测试
func runBatchTests(c *client.Client, folderID int) {
	fmt.Println("\n--- E. 批量操作 (batch) ---")

	// E1: 上传多个文件
	fmt.Println("[E1] 上传多个文件...")
	tmpFiles := make([]string, 3)
	fileIDs := make([]int, 3)
	for i := 0; i < 3; i++ {
		tmp, _ := createTempFile([]byte(fmt.Sprintf("batch file %d", i+1)))
		tmpFiles[i] = tmp
		defer os.Remove(tmp)
		info, _ := files.Upload(c, tmp, folderID)
		fileIDs[i] = parseFileID(info.ID)
	}
	fmt.Printf("  ✅ 上传 %d 个文件\n", len(fileIDs))

	// E2: 批量删除
	fmt.Println("[E2] 批量删除...")
	for _, fid := range fileIDs {
		files.DeleteFile(c, fid)
	}
	fmt.Printf("  ✅ 已删除 %d 个文件\n", len(fileIDs))

	// E3: 批量恢复 (恢复所有)
	fmt.Println("[E3] 恢复全部...")
	if err := recycle.RestoreAll(c); err != nil {
		fmt.Printf("  ⚠️  恢复全部失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ 已恢复全部\n")
	}

	// 检查恢复结果
	items, _ := recycle.ListRecycle(c, 1)
	fmt.Printf("  ℹ️  当前回收站 %d 项\n", len(items))
}

// runDangerousTests 执行危险操作测试
func runDangerousTests(c *client.Client) {
	fmt.Println("\n--- F. 危险操作 (clear) ---")

	// 确保有测试数据在回收站
	fmt.Println("ℹ️  上传文件准备危险测试...")
	tmpFile, _ := createTempFile([]byte("for dangerous test"))
	defer os.Remove(tmpFile)
	info, _ := files.Upload(c, tmpFile, -1)
	fileID := parseFileID(info.ID)
	files.DeleteFile(c, fileID)

	itemsBefore, _ := recycle.ListRecycle(c, 1)
	fmt.Printf("  ℹ️  当前回收站 %d 项\n", len(itemsBefore))

	// F1: 清空回收站
	fmt.Println("[F1] 清空回收站...")
	if err := recycle.ClearRecycle(c); err != nil {
		fmt.Printf("  ❌ 清空失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ 清空成功\n")
	}
}

// cleanup 执行清理
func cleanup(c *client.Client, folderID, subFolderID int) {
	fmt.Println("\n--- Cleanup ---")

	// 尝试删除测试文件夹（可能被清空了）
	if err := folders.DeleteFolder(c, folderID); err != nil {
		fmt.Printf("  ⚠️  清理失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ 已清理 folder_id=%d\n", folderID)
	}
}

// printStats 打印统计信息
func printStats() {
	fmt.Println("\n📊 使用说明:")
	fmt.Println("  基本测试:")
	fmt.Println("    ./lazgo_e2e.exe          # 使用已有 cookie")
	fmt.Println("\n账密登录测试:")
	fmt.Println("    ./lazgo_e2e.exe -login  # 用账号密码登录")
	fmt.Println("\n批量操作测试:")
	fmt.Println("    ./lazgo_e2e.exe -batch # +批量操作")
	fmt.Println("\n危险操作测试:")
	fmt.Println("    ./lazgo_e2e.exe -clear # +清空回收站")
	fmt.Println("\n完整测试:")
	fmt.Println("    ./lazgo_e2e.exe -login -batch -clear")
}

// printHelp 打印帮助
func printHelp() {
	fmt.Print(`lazgo E2E 完整测试工具

用法:
  lazgo_e2e.exe [选项]

选项:
  -login  使用账号密码登录测试 (默认使用已有 cookie)
  -batch  执行批量操作测试
  -clear  执行清空回收站测试 (危险操作!)
  -h      显示帮助

示例:
  lazgo_e2e.exe              # 基本测试
  lazgo_e2e.exe -login       # 登录测试
  lazgo_e2e.exe -batch      # 批量测试
  lazgo_e2e.exe -clear     # 包含清空
  lazgo_e2e.exe -login -batch -clear  # 全量测试
`)
}

// ==================== 工具函数 ====================

// safeDelay 随机延迟，防止请求过快被封号
func safeDelay(descr string) {
	// 随机 2-5 秒延迟
	sec := 2 + rand.Intn(4)
	fmt.Printf("  ⏳ %s (等待 %ds)...\n", descr, sec)
	time.Sleep(time.Duration(sec) * time.Second)
}

func parseFileID(s string) int {
	var id int
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			id = id*10 + int(ch-'0')
		}
	}
	return id
}