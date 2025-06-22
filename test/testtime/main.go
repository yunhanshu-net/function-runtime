package main

import (
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/yunhanshu-net/pkg/typex"
)

// 用户模型 - 使用typex.Time替代string
type User struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	BirthDate  typex.Time `json:"birth_date"`  // 生日
	CreateTime typex.Time `json:"create_time"` // 注册时间
	LastLogin  typex.Time `json:"last_login"`  // 最后登录
}

// 活动模型 - 展示时间运算的实际应用
type Event struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	StartTime typex.Time `json:"start_time"` // 开始时间
	EndTime   typex.Time `json:"end_time"`   // 结束时间
	CreateBy  int        `json:"create_by"`
}

func main() {
	fmt.Println("=== typex.Time 业务场景测试 ===")

	// 1. 创建用户数据
	userJSON := `{
		"id": 1,
		"name": "张三",
		"birth_date": "1990-05-15 00:00:00",
		"create_time": "2023-01-01 10:30:00",
		"last_login": "2025-06-13 09:15:30"
	}`

	var user User
	err := json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		fmt.Printf("用户数据解析错误: %v\n", err)
		return
	}

	fmt.Printf("用户: %s\n", user.Name)
	fmt.Printf("生日: %s\n", user.BirthDate.String())

	// 2. 年龄计算 - typex.Time的优势体现
	fmt.Println("\n=== 年龄计算 ===")
	birthTime := time.Time(user.BirthDate)
	age := time.Since(birthTime)
	years := int(age.Hours() / 24 / 365.25) // 考虑闰年
	fmt.Printf("年龄: %d岁\n", years)

	// 3. 账户活跃度分析
	fmt.Println("\n=== 账户活跃度分析 ===")
	createTime := time.Time(user.CreateTime)
	lastLoginTime := time.Time(user.LastLogin)

	// 注册天数
	registerDays := int(time.Since(createTime).Hours() / 24)
	fmt.Printf("注册天数: %d天\n", registerDays)

	// 最后登录距今天数
	lastLoginDays := int(time.Since(lastLoginTime).Hours() / 24)
	fmt.Printf("最后登录: %d天前\n", lastLoginDays)

	// 判断用户活跃度
	if lastLoginDays <= 7 {
		fmt.Println("用户状态: 活跃用户")
	} else if lastLoginDays <= 30 {
		fmt.Println("用户状态: 一般活跃")
	} else {
		fmt.Println("用户状态: 不活跃用户")
	}

	// 4. 活动时间管理
	fmt.Println("\n=== 活动时间管理 ===")
	eventJSON := `{
		"id": 1,
		"title": "产品发布会",
		"start_time": "2025-06-20 14:00:00",
		"end_time": "2025-06-20 17:00:00",
		"create_by": 1
	}`

	var event Event
	err = json.Unmarshal([]byte(eventJSON), &event)
	if err != nil {
		fmt.Printf("活动数据解析错误: %v\n", err)
		return
	}

	fmt.Printf("活动: %s\n", event.Title)

	// 活动时长计算
	startTime := time.Time(event.StartTime)
	endTime := time.Time(event.EndTime)
	duration := endTime.Sub(startTime)
	fmt.Printf("活动时长: %.0f小时\n", duration.Hours())

	// 距离活动开始时间
	now := time.Now()
	if startTime.After(now) {
		timeToStart := startTime.Sub(now)
		days := int(timeToStart.Hours() / 24)
		hours := int(timeToStart.Hours()) % 24
		fmt.Printf("距离开始: %d天%d小时\n", days, hours)
	} else if endTime.After(now) {
		fmt.Println("活动状态: 进行中")
	} else {
		fmt.Println("活动状态: 已结束")
	}

	// 5. 数据库查询示例（模拟）
	fmt.Println("\n=== 数据库查询示例 ===")

	// 查询最近30天注册的用户
	thirtyDaysAgo := typex.Time(time.Now().AddDate(0, 0, -30))
	fmt.Printf("查询条件: create_time > %s\n", thirtyDaysAgo.String())

	// 查询今天生日的用户
	today := time.Now()
	birthdayStart := typex.Time(time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()))
	birthdayEnd := typex.Time(time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, today.Location()))
	fmt.Printf("生日查询: birth_date BETWEEN %s AND %s\n", birthdayStart.String(), birthdayEnd.String())

	// 6. 时间戳和格式化
	fmt.Println("\n=== 时间戳和格式化 ===")
	fmt.Printf("Unix时间戳: %d\n", user.LastLogin.GetUnix())
	fmt.Printf("中文格式: %s\n", time.Time(user.BirthDate).Format("2006年01月02日"))
	fmt.Printf("ISO格式: %s\n", time.Time(user.CreateTime).Format(time.RFC3339))

	// 7. JSON序列化测试
	fmt.Println("\n=== JSON序列化测试 ===")
	userBytes, _ := json.Marshal(user)
	fmt.Printf("用户JSON: %s\n", string(userBytes))

	eventBytes, _ := json.Marshal(event)
	fmt.Printf("活动JSON: %s\n", string(eventBytes))

	fmt.Println("\n✅ typex.Time 完美支持所有时间运算场景！")
}
