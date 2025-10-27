package utils

import (
	"bytes"
	"strings"
	"testing"
)

func TestMarshalExt(t *testing.T) {
	// 测试数据
	type User struct {
		ID       int    `json:"id" xml:"id"`
		Name     string `json:"name" xml:"name"`
		Email    string `json:"email" xml:"email"`
		IsActive bool   `json:"is_active" xml:"is_active"`
	}

	user := User{
		ID:       1,
		Name:     "张三",
		Email:    "zhangsan@example.com",
		IsActive: true,
	}

	t.Run("JSONFormat", func(t *testing.T) {
		marshal := NewMarshalExt(MarshalOptions{Format: JSONFormat})

		// 测试序列化
		data, err := marshal.Marshal(user)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		// 测试反序列化
		var decoded User
		err = marshal.Unmarshal(data, &decoded)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if decoded.ID != user.ID || decoded.Name != user.Name {
			t.Error("Decoded data doesn't match original")
		}
	})

	t.Run("PrettyJSON", func(t *testing.T) {
		marshal := NewMarshalExt(MarshalOptions{
			Format: JSONFormat,
			Pretty: true,
			Indent: "  ",
		})

		data, err := marshal.Marshal(user)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		jsonStr := string(data)
		if !strings.Contains(jsonStr, "\n") {
			t.Error("Pretty JSON should contain newlines")
		}
	})

	t.Run("XMLFormat", func(t *testing.T) {
		marshal := NewMarshalExt(MarshalOptions{Format: XMLFormat})

		data, err := marshal.Marshal(user)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		xmlStr := string(data)
		if !strings.Contains(xmlStr, "<User>") {
			t.Error("XML should contain User tag")
		}
	})

	t.Run("StringFormat", func(t *testing.T) {
		marshal := NewMarshalExt(MarshalOptions{Format: StringFormat})

		// 测试字符串
		str := "hello world"
		data, err := marshal.Marshal(str)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		if string(data) != str {
			t.Errorf("Expected %s, got %s", str, string(data))
		}

		// 测试数字
		num := 123
		data, err = marshal.Marshal(num)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		if string(data) != "123" {
			t.Errorf("Expected 123, got %s", string(data))
		}
	})

	t.Run("MaxLength", func(t *testing.T) {
		marshal := NewMarshalExt(MarshalOptions{
			Format:    JSONFormat,
			MaxLength: 50,
			Truncate:  true,
		})

		data, err := marshal.Marshal(user)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		if len(data) > 50 {
			t.Errorf("Data length %d exceeds max length 50", len(data))
		}
	})
}

func TestMarshalExtChainMethods(t *testing.T) {
	user := map[string]interface{}{
		"id":   1,
		"name": "张三",
		"age":  25,
	}

	t.Run("ChainMethods", func(t *testing.T) {
		marshal := NewMarshalExt(DefaultMarshalOptions).
			SetFormat(JSONFormat).
			SetPretty(true).
			SetIndent("  ")

		data, err := marshal.Marshal(user)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		jsonStr := string(data)
		if !strings.Contains(jsonStr, "\n") {
			t.Error("Pretty JSON should contain newlines")
		}
	})

	t.Run("Clone", func(t *testing.T) {
		original := NewMarshalExt(MarshalOptions{
			Format: JSONFormat,
			Pretty: true,
		})

		cloned := original.Clone().SetFormat(XMLFormat)

		// 验证克隆不影响原始对象
		if original.options.Format != JSONFormat {
			t.Error("Original format should not be changed")
		}

		if cloned.options.Format != XMLFormat {
			t.Error("Cloned format should be XML")
		}
	})
}

func TestMarshalExtConvenienceMethods(t *testing.T) {
	user := map[string]interface{}{
		"id":   1,
		"name": "张三",
		"age":  25,
	}

	marshal := NewMarshalExt(DefaultMarshalOptions)

	t.Run("ToJSON", func(t *testing.T) {
		jsonStr, err := marshal.ToJSON(user)
		if err != nil {
			t.Fatalf("ToJSON failed: %v", err)
		}

		if !strings.Contains(jsonStr, "张三") {
			t.Error("JSON should contain name")
		}
	})

	t.Run("ToPrettyJSON", func(t *testing.T) {
		jsonStr, err := marshal.ToPrettyJSON(user)
		if err != nil {
			t.Fatalf("ToPrettyJSON failed: %v", err)
		}

		if !strings.Contains(jsonStr, "\n") {
			t.Error("Pretty JSON should contain newlines")
		}
	})

	t.Run("ToXML", func(t *testing.T) {
		xmlStr, err := marshal.ToXML(user)
		if err != nil {
			t.Fatalf("ToXML failed: %v", err)
		}

		// 对于 map 类型，XML 会回退到 JSON 格式
		if !strings.Contains(xmlStr, "张三") {
			t.Error("XML should contain name")
		}
	})

	t.Run("ToString", func(t *testing.T) {
		str, err := marshal.ToString(user)
		if err != nil {
			t.Fatalf("ToString failed: %v", err)
		}

		if str == "" {
			t.Error("String should not be empty")
		}
	})
}

func TestMarshalBuilder(t *testing.T) {
	user := map[string]interface{}{
		"id":   1,
		"name": "张三",
		"age":  25,
	}

	t.Run("BasicBuilder", func(t *testing.T) {
		builder := NewMarshalBuilder(user)

		jsonStr, err := builder.
			SetFormat(JSONFormat).
			SetPretty(true).
			BuildString()

		if err != nil {
			t.Fatalf("BuildString failed: %v", err)
		}

		if !strings.Contains(jsonStr, "张三") {
			t.Error("JSON should contain name")
		}
	})

	t.Run("FormatSpecificMethods", func(t *testing.T) {
		builder := NewMarshalBuilder(user)

		// 测试 ToJSON
		jsonStr, err := builder.ToJSON()
		if err != nil {
			t.Fatalf("ToJSON failed: %v", err)
		}

		if !strings.Contains(jsonStr, "张三") {
			t.Error("JSON should contain name")
		}

		// 测试 ToPrettyJSON
		prettyStr, err := builder.ToPrettyJSON()
		if err != nil {
			t.Fatalf("ToPrettyJSON failed: %v", err)
		}

		if !strings.Contains(prettyStr, "\n") {
			t.Error("Pretty JSON should contain newlines")
		}

		// 测试 ToXML
		xmlStr, err := builder.ToXML()
		if err != nil {
			t.Fatalf("ToXML failed: %v", err)
		}

		// 对于 map 类型，XML 会回退到 JSON 格式
		if !strings.Contains(xmlStr, "张三") {
			t.Error("XML should contain name")
		}
	})

	t.Run("BuilderClone", func(t *testing.T) {
		original := NewMarshalBuilder(user).SetPretty(true)
		cloned := original.Clone().SetFormat(XMLFormat)

		// 验证克隆不影响原始构建器
		originalStr, _ := original.ToJSON()
		clonedStr, _ := cloned.ToXML()

		if !strings.Contains(originalStr, "张三") {
			t.Error("Original should contain name")
		}

		// 对于 map 类型，XML 会回退到 JSON 格式
		if !strings.Contains(clonedStr, "张三") {
			t.Error("Cloned should contain name")
		}
	})
}

func TestTypeConverter(t *testing.T) {
	user := map[string]interface{}{
		"id":   1,
		"name": "张三",
		"age":  25,
	}

	t.Run("BasicConversion", func(t *testing.T) {
		converter := NewTypeConverter(user)

		// 测试 ToString
		str := converter.ToString()
		if str == "" {
			t.Error("String should not be empty")
		}

		// 测试 ToJSON
		jsonStr, err := converter.ToJSON()
		if err != nil {
			t.Fatalf("ToJSON failed: %v", err)
		}

		if !strings.Contains(jsonStr, "张三") {
			t.Error("JSON should contain name")
		}

		// 测试 ToPrettyJSON
		prettyStr, err := converter.ToPrettyJSON()
		if err != nil {
			t.Fatalf("ToPrettyJSON failed: %v", err)
		}

		if !strings.Contains(prettyStr, "\n") {
			t.Error("Pretty JSON should contain newlines")
		}
	})

	t.Run("DifferentTypes", func(t *testing.T) {
		// 测试字符串
		strConverter := NewTypeConverter("hello world")
		if strConverter.ToString() != "hello world" {
			t.Error("String conversion failed")
		}

		// 测试数字
		numConverter := NewTypeConverter(123)
		if numConverter.ToString() != "123" {
			t.Error("Number conversion failed")
		}

		// 测试布尔值
		boolConverter := NewTypeConverter(true)
		if boolConverter.ToString() != "true" {
			t.Error("Boolean conversion failed")
		}
	})
}

func TestGlobalFunctions(t *testing.T) {
	user := map[string]interface{}{
		"id":   1,
		"name": "张三",
		"age":  25,
	}

	t.Run("GlobalMarshalFunctions", func(t *testing.T) {
		// 测试 Marshal
		data, err := Marshal(user)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		if len(data) == 0 {
			t.Error("Marshal should return data")
		}

		// 测试 MarshalToString
		str, err := MarshalToString(user)
		if err != nil {
			t.Fatalf("MarshalToString failed: %v", err)
		}

		if !strings.Contains(str, "张三") {
			t.Error("String should contain name")
		}
	})

	t.Run("GlobalConversionFunctions", func(t *testing.T) {
		// 测试 ConvertToString
		str := ConvertToString(user)
		if str == "" {
			t.Error("ConvertToString should return non-empty string")
		}

		// 测试 ConvertToJSON
		jsonStr, err := ConvertToJSON(user)
		if err != nil {
			t.Fatalf("ConvertToJSON failed: %v", err)
		}

		if !strings.Contains(jsonStr, "张三") {
			t.Error("JSON should contain name")
		}

		// 测试 ConvertToPrettyJSON
		prettyStr, err := ConvertToPrettyJSON(user)
		if err != nil {
			t.Fatalf("ConvertToPrettyJSON failed: %v", err)
		}

		if !strings.Contains(prettyStr, "\n") {
			t.Error("Pretty JSON should contain newlines")
		}
	})

	t.Run("GlobalFormatFunctions", func(t *testing.T) {
		// 测试 ToJSON
		jsonStr, err := ToJSON(user)
		if err != nil {
			t.Fatalf("ToJSON failed: %v", err)
		}

		if !strings.Contains(jsonStr, "张三") {
			t.Error("JSON should contain name")
		}

		// 测试 ToPrettyJSON
		prettyStr, err := ToPrettyJSON(user)
		if err != nil {
			t.Fatalf("ToPrettyJSON failed: %v", err)
		}

		if !strings.Contains(prettyStr, "\n") {
			t.Error("Pretty JSON should contain newlines")
		}

		// 测试 ToXML
		xmlStr, err := ToXML(user)
		if err != nil {
			t.Fatalf("ToXML failed: %v", err)
		}

		// 对于 map 类型，XML 会回退到 JSON 格式
		if !strings.Contains(xmlStr, "张三") {
			t.Error("XML should contain name")
		}
	})
}

func TestMarshalExtWithWriter(t *testing.T) {
	user := map[string]interface{}{
		"id":   1,
		"name": "张三",
		"age":  25,
	}

	t.Run("MarshalToWriter", func(t *testing.T) {
		var buf bytes.Buffer
		marshal := NewMarshalExt(MarshalOptions{Format: JSONFormat})

		err := marshal.MarshalToWriter(&buf, user)
		if err != nil {
			t.Fatalf("MarshalToWriter failed: %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "张三") {
			t.Error("Writer should contain name")
		}
	})

	t.Run("BuilderToWriter", func(t *testing.T) {
		var buf bytes.Buffer
		builder := NewMarshalBuilder(user)

		err := builder.
			SetFormat(JSONFormat).
			SetPretty(true).
			BuildToWriter(&buf)

		if err != nil {
			t.Fatalf("BuildToWriter failed: %v", err)
		}

		result := buf.String()
		if !strings.Contains(result, "张三") {
			t.Error("Writer should contain name")
		}
	})
}

func TestMustMethods(t *testing.T) {
	user := map[string]interface{}{
		"id":   1,
		"name": "张三",
		"age":  25,
	}

	t.Run("MustMarshalToString", func(t *testing.T) {
		marshal := NewMarshalExt(DefaultMarshalOptions)

		// 测试正常情况
		result := marshal.MustMarshalToString(user)
		if !strings.Contains(result, "张三") {
			t.Error("MustMarshalToString should contain name")
		}

		// 测试 MustToJSON
		jsonStr := marshal.MustToJSON(user)
		if !strings.Contains(jsonStr, "张三") {
			t.Error("MustToJSON should contain name")
		}

		// 测试 MustToPrettyJSON
		prettyStr := marshal.MustToPrettyJSON(user)
		if !strings.Contains(prettyStr, "\n") {
			t.Error("MustToPrettyJSON should contain newlines")
		}
	})

	t.Run("MustMarshalBuilder", func(t *testing.T) {
		builder := NewMarshalBuilder(user)

		// 测试 MustBuildString
		result := builder.MustBuildString()
		if !strings.Contains(result, "张三") {
			t.Error("MustBuildString should contain name")
		}

		// 测试 MustToJSON
		jsonStr := builder.MustToJSON()
		if !strings.Contains(jsonStr, "张三") {
			t.Error("MustToJSON should contain name")
		}

		// 测试 MustToPrettyJSON
		prettyStr := builder.MustToPrettyJSON()
		if !strings.Contains(prettyStr, "\n") {
			t.Error("MustToPrettyJSON should contain newlines")
		}
	})

	t.Run("MustTypeConverter", func(t *testing.T) {
		converter := NewTypeConverter(user)

		// 测试 MustToJSON
		jsonStr := converter.MustToJSON()
		if !strings.Contains(jsonStr, "张三") {
			t.Error("MustToJSON should contain name")
		}

		// 测试 MustToPrettyJSON
		prettyStr := converter.MustToPrettyJSON()
		if !strings.Contains(prettyStr, "\n") {
			t.Error("MustToPrettyJSON should contain newlines")
		}
	})

	t.Run("MustGlobalFunctions", func(t *testing.T) {
		// 测试 MustMarshalToString
		result := MustMarshalToString(user)
		if !strings.Contains(result, "张三") {
			t.Error("MustMarshalToString should contain name")
		}

		// 测试 MustToJSON
		jsonStr := MustToJSON(user)
		if !strings.Contains(jsonStr, "张三") {
			t.Error("MustToJSON should contain name")
		}

		// 测试 MustToPrettyJSON
		prettyStr := MustToPrettyJSON(user)
		if !strings.Contains(prettyStr, "\n") {
			t.Error("MustToPrettyJSON should contain newlines")
		}

		// 测试 MustConvertToJSON
		convertStr := MustConvertToJSON(user)
		if !strings.Contains(convertStr, "张三") {
			t.Error("MustConvertToJSON should contain name")
		}
	})

	t.Run("MustReturnEmptyOnError", func(t *testing.T) {
		// 创建一个会导致序列化失败的对象
		invalidData := make(chan int)

		// 测试 MustMarshalToString 在错误时返回空字符串
		result := MustMarshalToString(invalidData)
		if result != "" {
			t.Error("MustMarshalToString should return empty string on error")
		}

		// 测试 MustToJSON 在错误时返回空字符串
		jsonResult := MustToJSON(invalidData)
		if jsonResult != "" {
			t.Error("MustToJSON should return empty string on error")
		}
	})
}

func BenchmarkMarshalExt(b *testing.B) {
	user := map[string]interface{}{
		"id":        1,
		"name":      "张三",
		"email":     "zhangsan@example.com",
		"is_active": true,
		"tags":      []string{"tag1", "tag2", "tag3"},
		"metadata": map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
	}

	b.Run("JSONMarshal", func(b *testing.B) {
		marshal := NewMarshalExt(MarshalOptions{Format: JSONFormat})
		for i := 0; i < b.N; i++ {
			_, err := marshal.Marshal(user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("PrettyJSONMarshal", func(b *testing.B) {
		marshal := NewMarshalExt(MarshalOptions{
			Format: JSONFormat,
			Pretty: true,
		})
		for i := 0; i < b.N; i++ {
			_, err := marshal.Marshal(user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("XMLMarshal", func(b *testing.B) {
		marshal := NewMarshalExt(MarshalOptions{Format: XMLFormat})
		for i := 0; i < b.N; i++ {
			_, err := marshal.Marshal(user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("StringMarshal", func(b *testing.B) {
		marshal := NewMarshalExt(MarshalOptions{Format: StringFormat})
		for i := 0; i < b.N; i++ {
			_, err := marshal.Marshal(user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MarshalBuilder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			builder := NewMarshalBuilder(user)
			_, err := builder.ToJSON()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("TypeConverter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			converter := NewTypeConverter(user)
			_, err := converter.ToJSON()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
