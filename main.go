package main

import (
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

	// 前往签到页面
	if _, err = page.Goto("https://bbs.topfeel.com/h5/#/minePages/qiandao",
		playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle}); err != nil {
		log.Fatalf("前往签到页面失败: %v", err)
	}
	time.Sleep(3 * time.Second)
	log.Println("已前往签到页面")

	// 检查是否未登录
	if visible, _ := page.Locator(".login-tips-block").IsVisible(); visible {
		log.Fatalln("未登录，无法签到")
	}
	log.Println("已登录网站，可进行签到")

	// 检查是否已签
	if visible, _ := page.Locator("uni-button:has(> .yiqian)").IsVisible(); visible {
		log.Println("今日已签到")
		return
	}
	log.Println("未签到，开始进行签到操作")

	// 加载签到按钮
	signButton := page.Locator(`uni-button:has(> .weiqian)`).First()
	err = signButton.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible})
	if err != nil {
		log.Fatalf("等待签到按钮出现失败：%s", err)
	}
	log.Println("已加载签到按钮")

	// 点击签到按钮
	if err = signButton.Click(); err != nil {
		log.Fatalf("点击签到按钮失败：%s", err)
	}
	log.Println("已点击签到按钮")

	// 等待滑块弹框出现
	err = page.Locator(`.zmm-slider-verify-title`).First().WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		log.Fatalf("等待滑块弹框出现失败：%s", err)
	}
	log.Println("已等待滑块弹框出现")

	// 获取滑动方块和验证方块的坐标
	touchBlock, err := page.Locator(`.zmm-slider-verify-block-touch`).First().BoundingBox()
	if err != nil {
		log.Fatalf("获取滑动方块坐标失败：%s", err)
	}
	verifyBlock, err := page.Locator(`.zmm-slider-verify-block-verify`).First().BoundingBox()
	if err != nil {
		log.Fatalf("获取验证方块坐标失败：%s", err)
	}
	log.Printf("滑动方块的坐标为：%f, %f", touchBlock.X, touchBlock.Y)
	log.Printf("验证方块的坐标为：%f, %f", verifyBlock.X, verifyBlock.Y)

	// 计算当前方块位置
	box := touchBlock
	centerX := box.X + box.Width/2
	centerY := box.Y + box.Height/2
	log.Printf("当前方块的中心坐标为：%f, %f", centerX, centerY)

	// 计算目标方块坐标
	targetX := verifyBlock.X + verifyBlock.Width/2
	targetY := touchBlock.Y + touchBlock.Height/2
	log.Printf("目标方块的中心坐标为：%f, %f", targetX, targetY)

	// 移动鼠标到当前方块中心坐标
	err = page.Mouse().Move(centerX, centerY)
	if err != nil {
		log.Fatalf("移动鼠标到当前方块中心坐标：%s", err)
	}
	log.Println("已移动鼠标到当前方块中心坐标")

	// 按下鼠标
	err = page.Mouse().Down()
	if err != nil {
		log.Fatalf("按下鼠标失败：%s", err)
	}
	log.Println("已按下鼠标")

	// 移动鼠标到目标方块中心坐标
	step := 20
	err = page.Mouse().Move(targetX, targetY, playwright.MouseMoveOptions{Steps: &step})
	if err != nil {
		log.Fatalf("移动鼠标到目标方块中心坐标失败：%s", err)
	}
	log.Println("已移动鼠标到目标方块中心坐标")

	// 抬起鼠标
	err = page.Mouse().Up()
	if err != nil {
		log.Fatalf("抬起鼠标失败：%s", err)
	}
	log.Println("已抬起鼠标")
}
