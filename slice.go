package utils

import (
	"reflect"
	"sort"
)

// Slice 切片操作器，提供链式调用的切片操作功能
type Slice[T any] struct {
	data []T
}

// NewSlice 创建一个新的切片操作器
func NewSlice[T any](data []T) *Slice[T] {
	return &Slice[T]{data: data}
}

// FromSlice 从现有切片创建切片操作器
func FromSlice[T any](slice []T) *Slice[T] {
	return &Slice[T]{data: slice}
}

// ToSlice 将操作器转换为切片
func (s *Slice[T]) ToSlice() []T {
	return s.data
}

// Len 返回切片长度
func (s *Slice[T]) Len() int {
	return len(s.data)
}

// IsEmpty 检查切片是否为空
func (s *Slice[T]) IsEmpty() bool {
	return len(s.data) == 0
}

// Filter 过滤切片，返回满足条件的元素
func (s *Slice[T]) Filter(predicate func(T) bool) *Slice[T] {
	var result []T
	for _, item := range s.data {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return &Slice[T]{data: result}
}

// Map 对切片中的每个元素应用函数，返回新的切片
func (s *Slice[T]) Map(mapper func(T) T) *Slice[T] {
	result := make([]T, len(s.data))
	for i, item := range s.data {
		result[i] = mapper(item)
	}
	return &Slice[T]{data: result}
}

// MapTo 对切片中的每个元素应用函数，转换为不同类型的切片
func MapTo[T, U any](s *Slice[T], mapper func(T) U) *Slice[U] {
	result := make([]U, len(s.data))
	for i, item := range s.data {
		result[i] = mapper(item)
	}
	return &Slice[U]{data: result}
}

// Reduce 将切片归约为单个值
func (s *Slice[T]) Reduce(initial T, reducer func(T, T) T) T {
	result := initial
	for _, item := range s.data {
		result = reducer(result, item)
	}
	return result
}

// Find 查找切片中第一个满足条件的元素
func (s *Slice[T]) Find(predicate func(T) bool) (T, bool) {
	for _, item := range s.data {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// FindIndex 查找切片中第一个满足条件的元素的索引
func (s *Slice[T]) FindIndex(predicate func(T) bool) int {
	for i, item := range s.data {
		if predicate(item) {
			return i
		}
	}
	return -1
}

// Contains 检查切片是否包含指定元素
func (s *Slice[T]) Contains(item T) bool {
	for _, v := range s.data {
		if reflect.DeepEqual(v, item) {
			return true
		}
	}
	return false
}

// IndexOf 返回元素在切片中的索引，如果不存在返回-1
func (s *Slice[T]) IndexOf(item T) int {
	for i, v := range s.data {
		if reflect.DeepEqual(v, item) {
			return i
		}
	}
	return -1
}

// Unique 去除切片中的重复元素（仅适用于comparable类型）
func Unique[T comparable](s *Slice[T]) *Slice[T] {
	keys := make(map[T]bool)
	var result []T
	for _, item := range s.data {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return &Slice[T]{data: result}
}

// Reverse 反转切片
func (s *Slice[T]) Reverse() *Slice[T] {
	result := make([]T, len(s.data))
	for i, item := range s.data {
		result[len(s.data)-1-i] = item
	}
	return &Slice[T]{data: result}
}

// Chunk 将切片分割成指定大小的块
func (s *Slice[T]) Chunk(size int) [][]T {
	if size <= 0 {
		return nil
	}

	var result [][]T
	for i := 0; i < len(s.data); i += size {
		end := i + size
		if end > len(s.data) {
			end = len(s.data)
		}
		result = append(result, s.data[i:end])
	}
	return result
}

// FlattenSlice 展平二维切片
func FlattenSlice[T any](s *Slice[[]T]) *Slice[T] {
	var result []T
	for _, subSlice := range s.data {
		result = append(result, subSlice...)
	}
	return &Slice[T]{data: result}
}

// Take 取切片的前n个元素
func (s *Slice[T]) Take(n int) *Slice[T] {
	if n <= 0 {
		return &Slice[T]{data: []T{}}
	}
	if n >= len(s.data) {
		return s
	}
	return &Slice[T]{data: s.data[:n]}
}

// Drop 跳过切片的前n个元素
func (s *Slice[T]) Drop(n int) *Slice[T] {
	if n <= 0 {
		return s
	}
	if n >= len(s.data) {
		return &Slice[T]{data: []T{}}
	}
	return &Slice[T]{data: s.data[n:]}
}

// TakeWhile 取满足条件的前缀元素
func (s *Slice[T]) TakeWhile(predicate func(T) bool) *Slice[T] {
	var result []T
	for _, item := range s.data {
		if predicate(item) {
			result = append(result, item)
		} else {
			break
		}
	}
	return &Slice[T]{data: result}
}

// DropWhile 跳过满足条件的前缀元素
func (s *Slice[T]) DropWhile(predicate func(T) bool) *Slice[T] {
	for i, item := range s.data {
		if !predicate(item) {
			return &Slice[T]{data: s.data[i:]}
		}
	}
	return &Slice[T]{data: []T{}}
}

// GroupBy 根据键函数对切片进行分组
func GroupBy[T any, K comparable](s *Slice[T], keyFunc func(T) K) map[K]*Slice[T] {
	groups := make(map[K]*Slice[T])
	for _, item := range s.data {
		key := keyFunc(item)
		if groups[key] == nil {
			groups[key] = &Slice[T]{data: []T{}}
		}
		groups[key].data = append(groups[key].data, item)
	}
	return groups
}

// Partition 将切片分割为满足条件和不满足条件的两个切片
func (s *Slice[T]) Partition(predicate func(T) bool) (*Slice[T], *Slice[T]) {
	var trueSlice, falseSlice []T
	for _, item := range s.data {
		if predicate(item) {
			trueSlice = append(trueSlice, item)
		} else {
			falseSlice = append(falseSlice, item)
		}
	}
	return &Slice[T]{data: trueSlice}, &Slice[T]{data: falseSlice}
}

// Sort 对切片进行排序（原地排序）
func (s *Slice[T]) Sort(less func(T, T) bool) *Slice[T] {
	sort.Slice(s.data, func(i, j int) bool {
		return less(s.data[i], s.data[j])
	})
	return s
}

// SortBy 根据键函数对切片进行排序
func SortBy[T any, K comparable](s *Slice[T], keyFunc func(T) K) *Slice[T] {
	sort.Slice(s.data, func(i, j int) bool {
		return reflect.ValueOf(keyFunc(s.data[i])).String() < reflect.ValueOf(keyFunc(s.data[j])).String()
	})
	return s
}

// First 获取切片的第一个元素
func (s *Slice[T]) First() (T, bool) {
	if len(s.data) == 0 {
		var zero T
		return zero, false
	}
	return s.data[0], true
}

// Last 获取切片的最后一个元素
func (s *Slice[T]) Last() (T, bool) {
	if len(s.data) == 0 {
		var zero T
		return zero, false
	}
	return s.data[len(s.data)-1], true
}

// Count 计算满足条件的元素数量
func (s *Slice[T]) Count(predicate func(T) bool) int {
	count := 0
	for _, item := range s.data {
		if predicate(item) {
			count++
		}
	}
	return count
}

// Any 检查是否有任何元素满足条件
func (s *Slice[T]) Any(predicate func(T) bool) bool {
	for _, item := range s.data {
		if predicate(item) {
			return true
		}
	}
	return false
}

// All 检查是否所有元素都满足条件
func (s *Slice[T]) All(predicate func(T) bool) bool {
	for _, item := range s.data {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// Concat 连接多个切片
func (s *Slice[T]) Concat(other *Slice[T]) *Slice[T] {
	result := make([]T, len(s.data)+len(other.data))
	copy(result, s.data)
	copy(result[len(s.data):], other.data)
	return &Slice[T]{data: result}
}

// Intersect 计算两个切片的交集（仅适用于comparable类型）
func Intersect[T comparable](s *Slice[T], other *Slice[T]) *Slice[T] {
	set := make(map[T]bool)
	for _, item := range s.data {
		set[item] = true
	}

	var result []T
	for _, item := range other.data {
		if set[item] {
			result = append(result, item)
			delete(set, item) // 避免重复
		}
	}
	return &Slice[T]{data: result}
}

// Difference 计算两个切片的差集（仅适用于comparable类型）
func Difference[T comparable](s *Slice[T], other *Slice[T]) *Slice[T] {
	set := make(map[T]bool)
	for _, item := range other.data {
		set[item] = true
	}

	var result []T
	for _, item := range s.data {
		if !set[item] {
			result = append(result, item)
		}
	}
	return &Slice[T]{data: result}
}

// Union 计算两个切片的并集（仅适用于comparable类型）
func Union[T comparable](s *Slice[T], other *Slice[T]) *Slice[T] {
	set := make(map[T]bool)
	var result []T

	for _, item := range s.data {
		if !set[item] {
			set[item] = true
			result = append(result, item)
		}
	}

	for _, item := range other.data {
		if !set[item] {
			set[item] = true
			result = append(result, item)
		}
	}

	return &Slice[T]{data: result}
}

// ForEach 对每个元素执行操作
func (s *Slice[T]) ForEach(action func(T)) {
	for _, item := range s.data {
		action(item)
	}
}

// ForEachWithIndex 对每个元素执行操作，包含索引
func (s *Slice[T]) ForEachWithIndex(action func(int, T)) {
	for i, item := range s.data {
		action(i, item)
	}
}

// 便捷函数，用于快速创建切片操作器

// SliceFilter 过滤切片，返回满足条件的元素
func SliceFilter[T any](slice []T, predicate func(T) bool) []T {
	return NewSlice(slice).Filter(predicate).ToSlice()
}

// SliceMap 对切片中的每个元素应用函数，返回新的切片
func SliceMap[T, U any](slice []T, mapper func(T) U) []U {
	return MapTo(NewSlice(slice), mapper).ToSlice()
}

// SliceReduce 将切片归约为单个值
func SliceReduce[T, U any](slice []T, initial U, reducer func(U, T) U) U {
	result := initial
	for _, item := range slice {
		result = reducer(result, item)
	}
	return result
}

// SliceFind 查找切片中第一个满足条件的元素
func SliceFind[T any](slice []T, predicate func(T) bool) (T, bool) {
	return NewSlice(slice).Find(predicate)
}

// SliceFindIndex 查找切片中第一个满足条件的元素的索引
func SliceFindIndex[T any](slice []T, predicate func(T) bool) int {
	return NewSlice(slice).FindIndex(predicate)
}

// SliceContains 检查切片是否包含指定元素
func SliceContains[T comparable](slice []T, item T) bool {
	return NewSlice(slice).Contains(item)
}

// SliceIndexOf 返回元素在切片中的索引，如果不存在返回-1
func SliceIndexOf[T comparable](slice []T, item T) int {
	return NewSlice(slice).IndexOf(item)
}

// SliceUnique 去除切片中的重复元素（便捷函数）
func SliceUnique[T comparable](slice []T) []T {
	return Unique(NewSlice(slice)).ToSlice()
}

// SliceReverse 反转切片
func SliceReverse[T any](slice []T) []T {
	return NewSlice(slice).Reverse().ToSlice()
}

// SliceChunk 将切片分割成指定大小的块
func SliceChunk[T any](slice []T, size int) [][]T {
	return NewSlice(slice).Chunk(size)
}

// SliceFlatten 展平二维切片
func SliceFlatten[T any](slice [][]T) []T {
	return FlattenSlice(NewSlice(slice)).ToSlice()
}

// SliceTake 取切片的前n个元素
func SliceTake[T any](slice []T, n int) []T {
	return NewSlice(slice).Take(n).ToSlice()
}

// SliceDrop 跳过切片的前n个元素
func SliceDrop[T any](slice []T, n int) []T {
	return NewSlice(slice).Drop(n).ToSlice()
}

// SliceTakeWhile 取满足条件的前缀元素
func SliceTakeWhile[T any](slice []T, predicate func(T) bool) []T {
	return NewSlice(slice).TakeWhile(predicate).ToSlice()
}

// SliceDropWhile 跳过满足条件的前缀元素
func SliceDropWhile[T any](slice []T, predicate func(T) bool) []T {
	return NewSlice(slice).DropWhile(predicate).ToSlice()
}

// SliceGroupBy 根据键函数对切片进行分组（便捷函数）
func SliceGroupBy[T any, K comparable](slice []T, keyFunc func(T) K) map[K][]T {
	groups := GroupBy(NewSlice(slice), keyFunc)
	result := make(map[K][]T)
	for k, v := range groups {
		result[k] = v.ToSlice()
	}
	return result
}

// SlicePartition 将切片分割为满足条件和不满足条件的两个切片
func SlicePartition[T any](slice []T, predicate func(T) bool) ([]T, []T) {
	trueSlice, falseSlice := NewSlice(slice).Partition(predicate)
	return trueSlice.ToSlice(), falseSlice.ToSlice()
}

// SliceSort 对切片进行排序（原地排序）
func SliceSort[T any](slice []T, less func(T, T) bool) {
	NewSlice(slice).Sort(less)
}

// SliceSortBy 根据键函数对切片进行排序（便捷函数）
func SliceSortBy[T any, K comparable](slice []T, keyFunc func(T) K) {
	SortBy(NewSlice(slice), keyFunc)
}

// SliceIsEmpty 检查切片是否为空
func SliceIsEmpty[T any](slice []T) bool {
	return NewSlice(slice).IsEmpty()
}

// SliceFirst 获取切片的第一个元素
func SliceFirst[T any](slice []T) (T, bool) {
	return NewSlice(slice).First()
}

// SliceLast 获取切片的最后一个元素
func SliceLast[T any](slice []T) (T, bool) {
	return NewSlice(slice).Last()
}

// SliceCount 计算满足条件的元素数量
func SliceCount[T any](slice []T, predicate func(T) bool) int {
	return NewSlice(slice).Count(predicate)
}

// SliceAny 检查是否有任何元素满足条件
func SliceAny[T any](slice []T, predicate func(T) bool) bool {
	return NewSlice(slice).Any(predicate)
}

// SliceAll 检查是否所有元素都满足条件
func SliceAll[T any](slice []T, predicate func(T) bool) bool {
	return NewSlice(slice).All(predicate)
}

// SliceConcat 连接多个切片
func SliceConcat[T any](slices ...[]T) []T {
	if len(slices) == 0 {
		return []T{}
	}
	result := NewSlice(slices[0])
	for i := 1; i < len(slices); i++ {
		result = result.Concat(NewSlice(slices[i]))
	}
	return result.ToSlice()
}

// SliceIntersect 计算两个切片的交集（便捷函数）
func SliceIntersect[T comparable](slice1, slice2 []T) []T {
	return Intersect(NewSlice(slice1), NewSlice(slice2)).ToSlice()
}

// SliceDifference 计算两个切片的差集（便捷函数）
func SliceDifference[T comparable](slice1, slice2 []T) []T {
	return Difference(NewSlice(slice1), NewSlice(slice2)).ToSlice()
}

// SliceUnion 计算两个切片的并集（便捷函数）
func SliceUnion[T comparable](slice1, slice2 []T) []T {
	return Union(NewSlice(slice1), NewSlice(slice2)).ToSlice()
}
