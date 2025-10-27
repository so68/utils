package utils

import (
	"math/rand"
	"time"
)

// RandomGenerator 随机数生成器结构体
type RandomGenerator struct {
	rng *rand.Rand
}

// NewRandomGenerator 创建新的随机数生成器
func NewRandomGenerator() *RandomGenerator {
	return &RandomGenerator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewRandomGeneratorWithSeed 使用指定种子创建随机数生成器
func NewRandomGeneratorWithSeed(seed int64) *RandomGenerator {
	return &RandomGenerator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Int 生成 [0, max) 范围内的随机整数
func (rg *RandomGenerator) Int(max int) int {
	if max <= 0 {
		return 0
	}
	return rg.rng.Intn(max)
}

// IntRange 生成 [min, max) 范围内的随机整数
func (rg *RandomGenerator) IntRange(min, max int) int {
	if min >= max {
		return min
	}
	return min + rg.rng.Intn(max-min)
}

// Int64 生成 [0, max) 范围内的随机int64
func (rg *RandomGenerator) Int64(max int64) int64 {
	if max <= 0 {
		return 0
	}
	return rg.rng.Int63n(max)
}

// Int64Range 生成 [min, max) 范围内的随机int64
func (rg *RandomGenerator) Int64Range(min, max int64) int64 {
	if min >= max {
		return min
	}
	return min + rg.rng.Int63n(max-min)
}

// Float64 生成 [0.0, 1.0) 范围内的随机浮点数
func (rg *RandomGenerator) Float64() float64 {
	return rg.rng.Float64()
}

// Float64Range 生成 [min, max) 范围内的随机浮点数
func (rg *RandomGenerator) Float64Range(min, max float64) float64 {
	if min >= max {
		return min
	}
	return min + rg.rng.Float64()*(max-min)
}

// Bool 生成随机布尔值
func (rg *RandomGenerator) Bool() bool {
	return rg.rng.Intn(2) == 1
}

// ChoiceString 从字符串切片中随机选择一个元素
func (rg *RandomGenerator) ChoiceString(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return slice[rg.rng.Intn(len(slice))]
}

// ChoiceInt 从整数切片中随机选择一个元素
func (rg *RandomGenerator) ChoiceInt(slice []int) int {
	if len(slice) == 0 {
		return 0
	}
	return slice[rg.rng.Intn(len(slice))]
}

// ShuffleString 随机打乱字符串切片
func (rg *RandomGenerator) ShuffleString(slice []string) {
	rg.rng.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

// ShuffleInt 随机打乱整数切片
func (rg *RandomGenerator) ShuffleInt(slice []int) {
	rg.rng.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

// String 生成指定长度的随机字符串
func (rg *RandomGenerator) String(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rg.rng.Intn(len(charset))]
	}
	return string(b)
}

// StringWithCharset 使用指定字符集生成随机字符串
func (rg *RandomGenerator) StringWithCharset(length int, charset string) string {
	if len(charset) == 0 {
		return ""
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rg.rng.Intn(len(charset))]
	}
	return string(b)
}

// WeightedChoiceString 根据权重随机选择字符串元素
func (rg *RandomGenerator) WeightedChoiceString(items []string, weights []float64) string {
	if len(items) == 0 {
		return ""
	}
	if len(items) != len(weights) {
		return items[rg.rng.Intn(len(items))]
	}

	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}

	if totalWeight <= 0 {
		return items[rg.rng.Intn(len(items))]
	}

	random := rg.rng.Float64() * totalWeight
	currentWeight := 0.0

	for i, w := range weights {
		currentWeight += w
		if random <= currentWeight {
			return items[i]
		}
	}

	return items[len(items)-1]
}

// UUID 生成简单的UUID格式字符串
func (rg *RandomGenerator) UUID() string {
	return rg.StringWithCharset(8, "0123456789abcdef") + "-" +
		rg.StringWithCharset(4, "0123456789abcdef") + "-" +
		rg.StringWithCharset(4, "0123456789abcdef") + "-" +
		rg.StringWithCharset(4, "0123456789abcdef") + "-" +
		rg.StringWithCharset(12, "0123456789abcdef")
}

// Bytes 生成指定长度的随机字节数组
func (rg *RandomGenerator) Bytes(length int) []byte {
	b := make([]byte, length)
	for i := range b {
		b[i] = byte(rg.rng.Intn(256))
	}
	return b
}
