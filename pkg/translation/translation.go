package translation

import (
	"fmt"
	"reflect"
	"regexp"
	//校验的核心

	//当你调用 c.ShouldBindJSON 时，Gin 内部就是通过这个包去调用校验器的。没有它，Gin 就不知道怎么去校验数据

	"github.com/gin-gonic/gin/binding"

	//它负责执行 required, min=5, email 这些规则逻辑。

	"github.com/go-playground/validator/v10"

	//翻译器必须先理解这个语言的“基本物理规则”（比如复数规则），才能进行后续的文字翻译

	//中文里数字怎么写？（1,000.00）

	//中文里日期怎么写？（2023年1月1日）

	//中文里有复数吗？（没有，1个苹果，2个苹果；英文有，1 apple, 2 apples）

	"github.com/go-playground/locales/en"

	"github.com/go-playground/locales/zh"
	// 新增：日语的基础规则
	"github.com/go-playground/locales/ja"

	//这是一个通用翻译引擎。它是一个空的机器，你给它塞入 locales（规则）和翻译文件（文案），它负责吐出最终的句子。

	//为什么要用：它是 validator 库指定的翻译引擎。validator 产生的错误对象，必须通过这个引擎才能转成文字。

	ut "github.com/go-playground/universal-translator"

	//这里面存的是成千上万条现成的报错文案。

	//内容示例：

	//required -> "{0} 为必填字段"

	//email -> "{0} 必须是一个有效的邮箱"

	//为什么要用：这是最省事的地方！ 如果没有这两个包，你需要自己手动把 validator 的几十种错误规则一条条写成中文。用了它，一行代码 zh_translations.RegisterDefaultTranslations 就全搞定了。

	en_translations "github.com/go-playground/validator/v10/translations/en"

	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	// 新增：日语的默认 Validator 翻译
	ja_translations "github.com/go-playground/validator/v10/translations/ja"

	//这是一个独立的、通用的 i18n 库。

	//为什么要用：前面的包都只管“参数校验错误”。但是你的系统里还有“登录失败”、“余额不足”、“系统繁忙”这些业务错误。这个库专门用来管理你写在 .toml 文件里的那些业务文案。

	"github.com/nicksnyder/go-i18n/v2/i18n"

	//TOML 文件解析器。因为你的翻译文件（active.zh.toml）是 TOML 格式的，Go 语言原生看不懂，需要这个包来解析文件内容。

	"github.com/pelletier/go-toml/v2"

	//提供标准的语言标签（Tag）。

	//为什么要用：它定义了什么是标准的 "zh-CN", "en-US"。用它是为了规范化，防止你手写字符串出错，同时 go-i18n 需要用它来匹配最合适的语言。

	"golang.org/x/text/language"

	"io/fs" // 注意引入这个标准库接口
)

// ==========================================1
// 1. 全局配置与嵌入文件
// ==========================================

var (
	I18nBundle *i18n.Bundle
	Uni        *ut.UniversalTranslator
)

// 定义支持的语言列表
var supportedLangs = []string{"zh", "en", "ja"}

// 定义 【Validator Tag】 -> 【TOML Key】 的映射关系
// 这样我们在代码里只需要维护这个 Map，不需要写死翻译内容
var validationMapping = map[string]string{
	"mobile":     "Valid_Mobile",
	"is_chinese": "Valid_IsChinese",
	"id_card":    "Valid_IDCard",

	// 👇 新增这一行！告诉程序：required 规则也要去 TOML 里找 Valid_Required
	"required": "Valid_Required",
}

// ==========================================
// 2. 初始化核心组件
// ==========================================

func InitComponents(localeFS fs.FS) {
	// --- A. 初始化 I18nBundle (加载 TOML) ---
	I18nBundle = i18n.NewBundle(language.Chinese)
	I18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// 加载所有嵌入的文件
	//注意路径：因为 localeFS 是在根目录 embed 的，所以文件路径依然是 "locales/xxx.toml"
	_, err := I18nBundle.LoadMessageFileFS(localeFS, "locales/active.zh.toml")
	if err != nil {
		panic(fmt.Errorf("加载中文包失败: %v", err))
	}
	_, err = I18nBundle.LoadMessageFileFS(localeFS, "locales/active.en.toml")
	if err != nil {
		panic(fmt.Errorf("加载英文包失败: %v", err))
	}

	_, err = I18nBundle.LoadMessageFileFS(localeFS, "locales/active.ja.toml")
	if err != nil {
		panic(fmt.Errorf("加载日语包失败: %v", err))
	}

	// --- B. 初始化 Validator ---
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		zhT := zh.New()
		enT := en.New()
		jaT := ja.New() // 1. 新建日语实例

		// 2. 把它塞进 Universal Translator
		// 第一个参数是 fallback (默认兜底)，后面的是支持的语言列表
		Uni = ut.New(enT, zhT, enT, jaT)

		// 1. 注册 TagNameFunc (核心技巧：使用占位符)
		// 优先取 label 标签，没有则取字段名
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := fld.Tag.Get("label")
			if name == "" {
				name = fld.Name
			}
			return "{" + name + "}"
		})

		// 2. 注册自定义校验规则
		_ = v.RegisterValidation("mobile", func(fl validator.FieldLevel) bool {
			ok, _ := regexp.MatchString(`^1[3-9]\d{9}$`, fl.Field().String())
			return ok
		})

		// (可以在这里继续添加 is_chinese, id_card 等规则)

		// 3. 循环注册翻译 (自动化逻辑：替代了手写的 registerTrans)
		// 遍历所有支持的语言 (zh, en)
		for _, lang := range supportedLangs {
			trans, found := Uni.GetTranslator(lang)
			if !found {
				continue
			}

			// 3.1 注册官方默认翻译 (处理 required, email 等)
			switch lang {
			case "zh":
				_ = zh_translations.RegisterDefaultTranslations(v, trans)
			case "en":
				_ = en_translations.RegisterDefaultTranslations(v, trans)
				// 新增：日语分支
			case "ja":
				_ = ja_translations.RegisterDefaultTranslations(v, trans)
			}

			// 3.2 注册自定义规则翻译 (从 TOML 读取)
			// 创建一个临时的 localizer 来读取该语言的配置
			localizer := i18n.NewLocalizer(I18nBundle, lang)

			for tag, tomlKey := range validationMapping {
				// 读取 TOML 中的文案，例如 Valid_Mobile
				msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: tomlKey})
				if err != nil {
					// 仅打印警告，不中断程序
					fmt.Printf("Warning: Missing translation for key '%s' in lang '%s'\n", tomlKey, lang)
					continue
				}

				// 注册到 Validator
				_ = v.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
					return ut.Add(tag, msg, true)
				}, func(ut ut.Translator, fe validator.FieldError) string {
					t, _ := ut.T(tag, fe.Field())
					return t
				})
			}
		}
	}
}
