package utils

import (
	"math"
	"testing"
)

/*
数字操作结构体功能测试

本文件用于测试Number结构体的各种功能特性，
包括基础数学运算、高级运算、比较操作、类型转换等。

运行命令：
go test -v -run "^TestNumber"

测试内容：
1. 基本数学运算 (Add, Sub, Mul, Div)
2. 高级数学运算 (Pow, Sqrt, Abs, Mod)
3. 取整运算 (Ceil, Floor, Round, RoundTo)
4. 比较运算 (GreaterThan, LessThan, Equal, IsPositive等)
5. 实用方法 (Max, Min, Neg, Set, Clone)
6. 类型转换 (Int, Int64, Float32)
7. 复杂链式调用测试
8. 边界条件和错误处理
9. 克隆功能独立性验证
10. 性能基准测试
*/

func TestNumberBasicOperations(t *testing.T) {
	// 测试基本运算
	n := NewNumber(10)

	// 链式调用测试: 10 + 5 = 15, 15 * 2 = 30, 30 - 3 = 27, 27 / 4 = 6.75
	result := n.Add(5).Mul(2).Sub(3).Div(4).Value()
	expected := 6.75
	if math.Abs(result-expected) > 1e-10 {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestNumberAdvancedOperations(t *testing.T) {
	n := NewNumber(16)

	// 测试幂运算和开方
	result := n.Pow(2).Sqrt().Value()
	expected := math.Sqrt(math.Pow(16, 2))
	if math.Abs(result-expected) > 1e-10 {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestNumberRounding(t *testing.T) {
	n := NewNumber(3.14159)

	// 测试四舍五入
	result := n.RoundTo(2).Value()
	expected := 3.14
	if math.Abs(result-expected) > 1e-10 {
		t.Errorf("Expected %f, got %f", expected, result)
	}

	// 测试向上取整
	n2 := NewNumber(3.2)
	result2 := n2.Ceil().Value()
	expected2 := 4.0
	if result2 != expected2 {
		t.Errorf("Expected %f, got %f", expected2, result2)
	}

	// 测试向下取整
	n3 := NewNumber(3.8)
	result3 := n3.Floor().Value()
	expected3 := 3.0
	if result3 != expected3 {
		t.Errorf("Expected %f, got %f", expected3, result3)
	}
}

func TestNumberComparisons(t *testing.T) {
	n := NewNumber(5)

	// 测试比较方法
	if !n.GreaterThan(3) {
		t.Error("Expected 5 > 3")
	}

	if !n.LessThan(7) {
		t.Error("Expected 5 < 7")
	}

	if !n.Equal(5) {
		t.Error("Expected 5 == 5")
	}

	if !n.IsPositive() {
		t.Error("Expected 5 to be positive")
	}

	if n.IsNegative() {
		t.Error("Expected 5 not to be negative")
	}

	if n.IsZero() {
		t.Error("Expected 5 not to be zero")
	}
}

func TestNumberUtilityMethods(t *testing.T) {
	n := NewNumber(-5)

	// 测试绝对值
	result := n.Abs().Value()
	expected := 5.0
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}

	// 测试最大值和最小值
	n2 := NewNumber(3)
	result2 := n2.Max(7).Value()
	expected2 := 7.0
	if result2 != expected2 {
		t.Errorf("Expected %f, got %f", expected2, result2)
	}

	n3 := NewNumber(10)
	result3 := n3.Min(5).Value()
	expected3 := 5.0
	if result3 != expected3 {
		t.Errorf("Expected %f, got %f", expected3, result3)
	}
}

func TestNumberClone(t *testing.T) {
	n1 := NewNumber(10)
	n2 := n1.Clone()

	// 修改 n1 不应该影响 n2
	n1.Add(5)
	if n2.Value() != 10 {
		t.Error("Clone should be independent")
	}

	if n1.Value() != 15 {
		t.Error("Original should be modified")
	}
}

func TestNumberTypeConversions(t *testing.T) {
	n := NewNumber(3.7)

	// 测试类型转换
	if n.Int() != 3 {
		t.Error("Int conversion failed")
	}

	if n.Int64() != 3 {
		t.Error("Int64 conversion failed")
	}

	if math.Abs(float64(n.Float32())-3.7) > 1e-6 {
		t.Error("Float32 conversion failed")
	}
}

func TestNumberComplexChain(t *testing.T) {
	// 复杂链式调用测试
	result := NewNumber(2).
		Add(3).  // 5
		Mul(4).  // 20
		Pow(2).  // 400
		Sqrt().  // 20
		Sub(5).  // 15
		Div(3).  // 5
		Round(). // 5
		Value()

	expected := 5.0
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestNumberEdgeCases(t *testing.T) {
	// 测试除零情况
	n := NewNumber(10)
	n.Div(0) // 应该不会改变原值
	if n.Value() != 10 {
		t.Error("Division by zero should not change value")
	}

	// 测试负数开方
	n2 := NewNumber(-4)
	n2.Sqrt() // 负数开方应该不改变原值
	if n2.Value() != -4 {
		t.Error("Square root of negative number should not change value")
	}

	// 测试取模运算
	n3 := NewNumber(10)
	n3.Mod(3)
	expected := 1.0
	if math.Abs(n3.Value()-expected) > 1e-10 {
		t.Errorf("Expected %f, got %f", expected, n3.Value())
	}
}

func BenchmarkNumberOperations(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewNumber(float64(i)).
			Add(1).
			Mul(2).
			Sub(1).
			Div(2).
			Value()
	}
}
