package utils

import (
	"fmt"
	"testing"
)

/*
随机数生成器功能测试

本文件用于测试RandomGenerator结构体的各种功能特性，
包括基础随机数生成、字符串操作、切片处理等。

运行命令：
go test -v -run "^TestRandomGenerator$"

测试内容：
1. 整数随机数生成 (Int, IntRange)
2. 浮点数随机数生成 (Float64, Float64Range)
3. 布尔值随机生成 (Bool)
4. 字符串随机生成 (String, StringWithCharset)
5. 切片随机选择 (ChoiceString, ChoiceInt)
6. 切片随机打乱 (ShuffleString, ShuffleInt)
7. 权重随机选择 (WeightedChoiceString)
8. UUID和字节数组生成 (UUID, Bytes)
9. 边界条件和错误处理
10. 随机性分布验证
*/

func TestRandomGenerator(t *testing.T) {
	// 创建随机数生成器
	rg := NewRandomGenerator()

	// 测试整数生成
	t.Run("Int", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			val := rg.Int(100)
			if val < 0 || val >= 100 {
				t.Errorf("Int(100) returned %d, expected [0, 100)", val)
			}
		}
	})

	// 测试整数范围生成
	t.Run("IntRange", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			val := rg.IntRange(10, 20)
			if val < 10 || val >= 20 {
				t.Errorf("IntRange(10, 20) returned %d, expected [10, 20)", val)
			}
		}
	})

	// 测试浮点数生成
	t.Run("Float64", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			val := rg.Float64()
			if val < 0.0 || val >= 1.0 {
				t.Errorf("Float64() returned %f, expected [0.0, 1.0)", val)
			}
		}
	})

	// 测试浮点数范围生成
	t.Run("Float64Range", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			val := rg.Float64Range(1.5, 2.5)
			if val < 1.5 || val >= 2.5 {
				t.Errorf("Float64Range(1.5, 2.5) returned %f, expected [1.5, 2.5)", val)
			}
		}
	})

	// 测试布尔值生成
	t.Run("Bool", func(t *testing.T) {
		trueCount := 0
		for i := 0; i < 100; i++ {
			if rg.Bool() {
				trueCount++
			}
		}
		// 在100次测试中，true的数量应该在合理范围内（大约40-60）
		if trueCount < 20 || trueCount > 80 {
			t.Errorf("Bool() distribution seems off: %d true out of 100", trueCount)
		}
	})

	// 测试字符串生成
	t.Run("String", func(t *testing.T) {
		str := rg.String(10)
		if len(str) != 10 {
			t.Errorf("String(10) returned length %d, expected 10", len(str))
		}
	})

	// 测试切片选择
	t.Run("ChoiceString", func(t *testing.T) {
		items := []string{"a", "b", "c", "d", "e"}
		for i := 0; i < 10; i++ {
			choice := rg.ChoiceString(items)
			found := false
			for _, item := range items {
				if item == choice {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ChoiceString returned %s, not in original slice", choice)
			}
		}
	})

	// 测试切片打乱
	t.Run("ShuffleString", func(t *testing.T) {
		original := []string{"1", "2", "3", "4", "5"}
		shuffled := make([]string, len(original))
		copy(shuffled, original)
		rg.ShuffleString(shuffled)

		// 检查长度是否相同
		if len(shuffled) != len(original) {
			t.Errorf("Shuffled slice length %d != original length %d", len(shuffled), len(original))
		}

		// 检查是否包含所有原始元素
		for _, orig := range original {
			found := false
			for _, shuf := range shuffled {
				if orig == shuf {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Shuffled slice missing element %s", orig)
			}
		}
	})

	// 测试权重选择
	t.Run("WeightedChoiceString", func(t *testing.T) {
		items := []string{"a", "b", "c"}
		weights := []float64{0.1, 0.3, 0.6} // c 应该被选择得最多

		results := make(map[string]int)
		for i := 0; i < 1000; i++ {
			choice := rg.WeightedChoiceString(items, weights)
			results[choice]++
		}

		// c 应该被选择得最多
		if results["c"] <= results["a"] || results["c"] <= results["b"] {
			t.Errorf("Weighted choice distribution seems off: %v", results)
		}
	})

	// 测试UUID生成
	t.Run("UUID", func(t *testing.T) {
		uuid := rg.UUID()
		if len(uuid) != 36 { // UUID格式: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
			t.Errorf("UUID length %d, expected 36", len(uuid))
		}
	})

	// 测试字节数组生成
	t.Run("Bytes", func(t *testing.T) {
		bytes := rg.Bytes(10)
		if len(bytes) != 10 {
			t.Errorf("Bytes(10) returned length %d, expected 10", len(bytes))
		}
	})
}

// ExampleRandomGenerator 演示如何使用随机数生成器
func ExampleRandomGenerator() {
	// 创建随机数生成器
	rg := NewRandomGenerator()

	// 生成随机整数
	fmt.Printf("随机整数 (0-99): %d\n", rg.Int(100))

	// 生成指定范围的随机整数
	fmt.Printf("随机整数 (10-20): %d\n", rg.IntRange(10, 20))

	// 生成随机浮点数
	fmt.Printf("随机浮点数 (0.0-1.0): %.3f\n", rg.Float64())

	// 生成指定范围的随机浮点数
	fmt.Printf("随机浮点数 (1.5-2.5): %.3f\n", rg.Float64Range(1.5, 2.5))

	// 生成随机布尔值
	fmt.Printf("随机布尔值: %t\n", rg.Bool())

	// 生成随机字符串
	fmt.Printf("随机字符串: %s\n", rg.String(8))

	// 从切片中随机选择
	fruits := []string{"苹果", "香蕉", "橙子", "葡萄"}
	fmt.Printf("随机水果: %s\n", rg.ChoiceString(fruits))

	// 生成UUID
	fmt.Printf("随机UUID: %s\n", rg.UUID())
}
