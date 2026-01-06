package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/playwright-community/playwright-go"
)

// 定义一些常量
var (
	SESSION_ID              string
	ACCESS_TOKEN            string
	ENABLE_BROWSER_HEADLESS bool
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 解析 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading: %v", err)
	}

	// 获取环境变量
	SESSION_ID = os.Getenv("PHP_SESSION_ID")
	ACCESS_TOKEN = os.Getenv("ACCESS_TOKEN")
	ENABLE_BROWSER_HEADLESS, _ = strconv.ParseBool(os.Getenv("ENABLE_BROWSER_HEADLESS"))
}

func main() {
	// 安装浏览器
	if err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}}); err != nil {
		log.Fatalf("浏览器安装失败: %v", err)
	}
	log.Println("浏览器已安装")

	// 启动 Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Playwright 启动失败: %v", err)
	}
	defer pw.Stop()
	log.Println("Playwright 已启动")

	// 创建浏览器实例
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(ENABLE_BROWSER_HEADLESS),
	})
	if err != nil {
		log.Fatalf("浏览器启动失败: %v", err)
	}
	defer browser.Close()
	log.Println("浏览器已启动")

	// 创建页面上下文
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{Width: 1920, Height: 1080},
		StorageState: &playwright.OptionalStorageState{
			Cookies: []playwright.OptionalCookie{{
				Name:   "PHPSESSID",
				Value:  SESSION_ID,
				Domain: playwright.String(".topfeel.com"),
				Path:   playwright.String("/"),
			}},
			Origins: []playwright.Origin{{
				Origin: "https://bbs.topfeel.com",
				LocalStorage: []playwright.NameValue{{
					Name:  "token",
					Value: ACCESS_TOKEN,
				}},
			}},
		},
	})
	if err != nil {
		log.Fatalf("上下文创建失败: %v", err)
	}
	defer context.Close()
	log.Println("页面上下文已创建")

	// 创建页面实例
	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("页面创建失败: %v", err)
	}
	log.Println("页面已创建")

	// 自动签到
	if err := autoSignIn(page); err != nil {
		log.Printf("签到失败: %v（程序继续运行）", err)
	} else {
		log.Println("签到完成")
	}

	// 自动评论
	if err := autoComment(page); err != nil {
		log.Printf("评论失败: %v（程序继续运行）", err)
	} else {
		log.Println("评论完成")
	}

	log.Println("程序已结束")
}

func autoSignIn(page playwright.Page) error {
	// 前往签到页面
	if _, err := page.Goto("https://bbs.topfeel.com/h5/#/minePages/qiandao",
		playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle}); err != nil {
		return fmt.Errorf("前往签到页面失败: %w", err)
	}
	time.Sleep(3 * time.Second)
	log.Println("已进入签到页面")

	// 检查是否未登录
	if visible, _ := page.Locator(".login-tips-block").IsVisible(); visible {
		return fmt.Errorf("未登录，无法签到")
	}
	log.Println("已登录，即将进行签到")

	// 检查是否已签到
	if visible, _ := page.Locator("uni-button:has(> .yiqian)").IsVisible(); visible {
		log.Println("今日已签到")
		return nil
	}
	log.Println("未签到，即将进行签到")

	// 加载并点击签到按钮
	signButton := page.Locator(`uni-button:has(> .weiqian)`).First()
	if err := signButton.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		return fmt.Errorf("等待签到按钮失败: %w", err)
	}
	if err := signButton.Click(); err != nil {
		return fmt.Errorf("点击签到按钮失败: %w", err)
	}
	log.Println("已点击签到按钮")

	// 等待滑块并完成验证
	if err := page.Locator(`.zmm-slider-verify-title`).First().WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	}); err != nil {
		return fmt.Errorf("等待滑块弹框失败: %w", err)
	}
	log.Println("已弹出滑块弹框")

	touchBlock, err := page.Locator(`.zmm-slider-verify-block-touch`).First().BoundingBox()
	if err != nil {
		return fmt.Errorf("获取滑动方块坐标失败: %w", err)
	}
	verifyBlock, err := page.Locator(`.zmm-slider-verify-block-verify`).First().BoundingBox()
	if err != nil {
		return fmt.Errorf("获取验证方块坐标失败: %w", err)
	}
	log.Println("已出现滑动方块和验证方块")

	centerX := touchBlock.X + touchBlock.Width/2
	centerY := touchBlock.Y + touchBlock.Height/2
	targetX := verifyBlock.X + verifyBlock.Width/2
	targetY := touchBlock.Y + touchBlock.Height/2

	mouse := page.Mouse()
	if err := mouse.Move(centerX, centerY); err != nil {
		return fmt.Errorf("移动鼠标到起始位置失败: %w", err)
	}
	if err := mouse.Down(); err != nil {
		return fmt.Errorf("按下鼠标失败: %w", err)
	}
	step := 20
	if err := mouse.Move(targetX, targetY, playwright.MouseMoveOptions{Steps: &step}); err != nil {
		return fmt.Errorf("滑动到目标位置失败: %w", err)
	}
	if err := mouse.Up(); err != nil {
		return fmt.Errorf("抬起鼠标失败: %w", err)
	}
	log.Println("已滑动滑块")

	return nil
}

func autoComment(page playwright.Page) error {
	log.Println("即将进行自动评论")

	return nil
}
