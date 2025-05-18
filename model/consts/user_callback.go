package consts

const (
	// 页面事件
	CallbackTypeOnPageLoad = "OnPageLoad" // 页面加载时

	// API 生命周期
	CallbackTypeOnApiCreated    = "OnApiCreated"    // API创建完成时
	CallbackTypeOnApiUpdated    = "OnApiUpdated"    // API更新时
	CallbackTypeBeforeApiDelete = "BeforeApiDelete" // API删除前
	CallbackTypeAfterApiDeleted = "AfterApiDeleted" // API删除后

	// 运行器(Runner)生命周期
	CallbackTypeBeforeRunnerClose = "BeforeRunnerClose" // 运行器关闭前
	CallbackTypeAfterRunnerClose  = "AfterRunnerClose"  // 运行器关闭后

	// 版本控制
	CallbackTypeOnVersionChange = "OnVersionChange" // 版本变更时

	// 输入交互
	CallbackTypeOnInputFuzzy    = "OnInputFuzzy"    // 输入模糊匹配
	CallbackTypeOnInputValidate = "OnInputValidate" // 输入校验

	// 表格操作
	CallbackTypeOnTableDeleteRows = "OnTableDeleteRows" // 删除表格行
	CallbackTypeOnTableUpdateRow  = "OnTableUpdateRow"  // 更新表格行
	CallbackTypeOnTableSearch     = "OnTableSearch"     // 表格搜索
)
