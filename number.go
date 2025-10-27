package utils

import (
	"math"
)

// Number 数字操作结构体，支持链式调用
type Number struct {
	value float64
}

// NewNumber 创建一个新的 Number 实例
func NewNumber(value float64) *Number {
	return &Number{value: value}
}

// Value 获取当前数值
func (n *Number) Value() float64 {
	return n.value
}

// Int 获取整数值
func (n *Number) Int() int {
	return int(n.value)
}

// Int64 获取 int64 值
func (n *Number) Int64() int64 {
	return int64(n.value)
}

// Float32 获取 float32 值
func (n *Number) Float32() float32 {
	return float32(n.value)
}

// Add 加法运算
func (n *Number) Add(value float64) *Number {
	n.value += value
	return n
}

// Sub 减法运算
func (n *Number) Sub(value float64) *Number {
	n.value -= value
	return n
}

// Mul 乘法运算
func (n *Number) Mul(value float64) *Number {
	n.value *= value
	return n
}

// Div 除法运算
func (n *Number) Div(value float64) *Number {
	if value != 0 {
		n.value /= value
	}
	return n
}

// Pow 幂运算
func (n *Number) Pow(exponent float64) *Number {
	n.value = math.Pow(n.value, exponent)
	return n
}

// Sqrt 开平方根
func (n *Number) Sqrt() *Number {
	if n.value >= 0 {
		n.value = math.Sqrt(n.value)
	}
	return n
}

// Abs 取绝对值
func (n *Number) Abs() *Number {
	n.value = math.Abs(n.value)
	return n
}

// Ceil 向上取整
func (n *Number) Ceil() *Number {
	n.value = math.Ceil(n.value)
	return n
}

// Floor 向下取整
func (n *Number) Floor() *Number {
	n.value = math.Floor(n.value)
	return n
}

// Round 四舍五入
func (n *Number) Round() *Number {
	n.value = math.Round(n.value)
	return n
}

// RoundTo 四舍五入到指定小数位
func (n *Number) RoundTo(places int) *Number {
	multiplier := math.Pow(10, float64(places))
	n.value = math.Round(n.value*multiplier) / multiplier
	return n
}

// Max 与另一个值比较取最大值
func (n *Number) Max(value float64) *Number {
	n.value = math.Max(n.value, value)
	return n
}

// Min 与另一个值比较取最小值
func (n *Number) Min(value float64) *Number {
	n.value = math.Min(n.value, value)
	return n
}

// Mod 取模运算
func (n *Number) Mod(value float64) *Number {
	if value != 0 {
		n.value = math.Mod(n.value, value)
	}
	return n
}

// Neg 取负数
func (n *Number) Neg() *Number {
	n.value = -n.value
	return n
}

// Set 设置新值
func (n *Number) Set(value float64) *Number {
	n.value = value
	return n
}

// Clone 克隆当前 Number 实例
func (n *Number) Clone() *Number {
	return &Number{value: n.value}
}

// IsZero 判断是否为零
func (n *Number) IsZero() bool {
	return n.value == 0
}

// IsPositive 判断是否为正数
func (n *Number) IsPositive() bool {
	return n.value > 0
}

// IsNegative 判断是否为负数
func (n *Number) IsNegative() bool {
	return n.value < 0
}

// IsInteger 判断是否为整数
func (n *Number) IsInteger() bool {
	return n.value == math.Trunc(n.value)
}

// Equal 判断是否等于指定值
func (n *Number) Equal(value float64) bool {
	return n.value == value
}

// GreaterThan 判断是否大于指定值
func (n *Number) GreaterThan(value float64) bool {
	return n.value > value
}

// LessThan 判断是否小于指定值
func (n *Number) LessThan(value float64) bool {
	return n.value < value
}

// GreaterOrEqual 判断是否大于等于指定值
func (n *Number) GreaterOrEqual(value float64) bool {
	return n.value >= value
}

// LessOrEqual 判断是否小于等于指定值
func (n *Number) LessOrEqual(value float64) bool {
	return n.value <= value
}
