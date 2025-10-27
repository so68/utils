package utils

import (
	"reflect"
	"testing"
)

/*
切片操作功能测试

本文件用于测试Slice结构体的各种功能特性，
包括基础切片操作、链式调用、函数式编程等。

运行命令：
go test -v -run "^TestSlice"

测试内容：
1. 基础操作 (NewSlice, Len, IsEmpty, ToSlice)
2. 过滤和映射 (Filter, Map, MapTo)
3. 归约和查找 (Reduce, Find, FindIndex)
4. 包含和索引 (Contains, IndexOf)
5. 去重和反转 (Unique, Reverse)
6. 分块和展平 (Chunk, FlattenSlice)
7. 取和丢弃 (Take, Drop, TakeWhile, DropWhile)
8. 分组和分割 (GroupBy, Partition)
9. 排序 (Sort, SortBy)
10. 首尾元素 (First, Last)
11. 计数和判断 (Count, Any, All)
12. 集合操作 (Concat, Intersect, Difference, Union)
13. 遍历操作 (ForEach, ForEachWithIndex)
14. 便捷函数测试
*/

// 测试结构体方式的链式调用
func TestSliceStruct(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6}

	// 测试链式调用
	result := NewSlice(numbers).
		Filter(func(n int) bool { return n%2 == 0 }).
		Map(func(n int) int { return n * n }).
		Take(2).
		ToSlice()

	expected := []int{4, 16} // 2*2, 4*4
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Chain operations = %v, want %v", result, expected)
	}
}

func TestNewSlice(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)

	if slice.Len() != 5 {
		t.Errorf("Len() = %v, want 5", slice.Len())
	}

	if slice.IsEmpty() {
		t.Errorf("IsEmpty() should be false")
	}
}

func TestSliceFilter(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6}
	slice := NewSlice(numbers)
	evenSlice := slice.Filter(func(n int) bool {
		return n%2 == 0
	})
	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(evenSlice.ToSlice(), expected) {
		t.Errorf("Slice.Filter() = %v, want %v", evenSlice.ToSlice(), expected)
	}
}

func TestSliceMap(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	squares := slice.Map(func(n int) int {
		return n * n
	})
	expected := []int{1, 4, 9, 16, 25}
	if !reflect.DeepEqual(squares.ToSlice(), expected) {
		t.Errorf("Slice.Map() = %v, want %v", squares.ToSlice(), expected)
	}
}

func TestMapTo(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	strings := MapTo(slice, func(n int) string {
		return string(rune('0' + n))
	})
	expected := []string{"1", "2", "3", "4", "5"}
	if !reflect.DeepEqual(strings.ToSlice(), expected) {
		t.Errorf("MapTo() = %v, want %v", strings.ToSlice(), expected)
	}
}

func TestSliceReduce(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	sum := slice.Reduce(0, func(acc, n int) int {
		return acc + n
	})
	expected := 15
	if sum != expected {
		t.Errorf("Slice.Reduce() = %v, want %v", sum, expected)
	}
}

func TestSliceFind(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	found, exists := slice.Find(func(n int) bool {
		return n > 3
	})
	if !exists || found != 4 {
		t.Errorf("Slice.Find() = %v, %v, want 4, true", found, exists)
	}

	_, exists = slice.Find(func(n int) bool {
		return n > 10
	})
	if exists {
		t.Errorf("Slice.Find() should not find element > 10")
	}
}

func TestSliceFindIndex(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	index := slice.FindIndex(func(n int) bool {
		return n == 3
	})
	if index != 2 {
		t.Errorf("Slice.FindIndex() = %v, want 2", index)
	}

	index = slice.FindIndex(func(n int) bool {
		return n == 10
	})
	if index != -1 {
		t.Errorf("Slice.FindIndex() = %v, want -1", index)
	}
}

func TestSliceContains(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	if !slice.Contains(3) {
		t.Errorf("Slice.Contains() should find 3")
	}
	if slice.Contains(6) {
		t.Errorf("Slice.Contains() should not find 6")
	}
}

func TestSliceIndexOf(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	index := slice.IndexOf(3)
	if index != 2 {
		t.Errorf("Slice.IndexOf() = %v, want 2", index)
	}

	index = slice.IndexOf(6)
	if index != -1 {
		t.Errorf("Slice.IndexOf() = %v, want -1", index)
	}
}

func TestSliceUnique(t *testing.T) {
	numbers := []int{1, 2, 2, 3, 3, 3, 4}
	slice := NewSlice(numbers)
	unique := Unique(slice)
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(unique.ToSlice(), expected) {
		t.Errorf("Slice.Unique() = %v, want %v", unique.ToSlice(), expected)
	}
}

func TestSliceReverse(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	reversed := slice.Reverse()
	expected := []int{5, 4, 3, 2, 1}
	if !reflect.DeepEqual(reversed.ToSlice(), expected) {
		t.Errorf("Slice.Reverse() = %v, want %v", reversed.ToSlice(), expected)
	}
}

func TestSliceChunk(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	slice := NewSlice(numbers)
	chunks := slice.Chunk(3)
	expected := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	if !reflect.DeepEqual(chunks, expected) {
		t.Errorf("Slice.Chunk() = %v, want %v", chunks, expected)
	}

	// Test with remainder
	numbers = []int{1, 2, 3, 4, 5}
	slice = NewSlice(numbers)
	chunks = slice.Chunk(2)
	expected = [][]int{{1, 2}, {3, 4}, {5}}
	if !reflect.DeepEqual(chunks, expected) {
		t.Errorf("Slice.Chunk() = %v, want %v", chunks, expected)
	}
}

func TestSliceFlatten(t *testing.T) {
	matrix := [][]int{{1, 2}, {3, 4}, {5, 6}}
	slice := NewSlice(matrix)
	flattened := FlattenSlice(slice)
	expected := []int{1, 2, 3, 4, 5, 6}
	if !reflect.DeepEqual(flattened.ToSlice(), expected) {
		t.Errorf("Slice.Flatten() = %v, want %v", flattened.ToSlice(), expected)
	}
}

func TestSliceTake(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	taken := slice.Take(3)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(taken.ToSlice(), expected) {
		t.Errorf("Slice.Take() = %v, want %v", taken.ToSlice(), expected)
	}

	// Test taking more than available
	taken = slice.Take(10)
	if !reflect.DeepEqual(taken.ToSlice(), numbers) {
		t.Errorf("Slice.Take() = %v, want %v", taken.ToSlice(), numbers)
	}
}

func TestSliceDrop(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	dropped := slice.Drop(2)
	expected := []int{3, 4, 5}
	if !reflect.DeepEqual(dropped.ToSlice(), expected) {
		t.Errorf("Slice.Drop() = %v, want %v", dropped.ToSlice(), expected)
	}

	// Test dropping more than available
	dropped = slice.Drop(10)
	expected = []int{}
	if !reflect.DeepEqual(dropped.ToSlice(), expected) {
		t.Errorf("Slice.Drop() = %v, want %v", dropped.ToSlice(), expected)
	}
}

func TestSliceTakeWhile(t *testing.T) {
	numbers := []int{2, 4, 6, 7, 8, 9}
	slice := NewSlice(numbers)
	taken := slice.TakeWhile(func(n int) bool {
		return n%2 == 0
	})
	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(taken.ToSlice(), expected) {
		t.Errorf("Slice.TakeWhile() = %v, want %v", taken.ToSlice(), expected)
	}
}

func TestSliceDropWhile(t *testing.T) {
	numbers := []int{2, 4, 6, 7, 8, 9}
	slice := NewSlice(numbers)
	dropped := slice.DropWhile(func(n int) bool {
		return n%2 == 0
	})
	expected := []int{7, 8, 9}
	if !reflect.DeepEqual(dropped.ToSlice(), expected) {
		t.Errorf("Slice.DropWhile() = %v, want %v", dropped.ToSlice(), expected)
	}
}

func TestSliceGroupBy(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	people := []Person{
		{"Alice", 25},
		{"Bob", 30},
		{"Charlie", 25},
		{"David", 30},
	}

	slice := NewSlice(people)
	groups := GroupBy(slice, func(p Person) int {
		return p.Age
	})

	expected25 := []Person{{"Alice", 25}, {"Charlie", 25}}
	expected30 := []Person{{"Bob", 30}, {"David", 30}}

	if !reflect.DeepEqual(groups[25].ToSlice(), expected25) {
		t.Errorf("Slice.GroupBy() age 25 = %v, want %v", groups[25].ToSlice(), expected25)
	}
	if !reflect.DeepEqual(groups[30].ToSlice(), expected30) {
		t.Errorf("Slice.GroupBy() age 30 = %v, want %v", groups[30].ToSlice(), expected30)
	}
}

func TestSlicePartition(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6}
	slice := NewSlice(numbers)
	even, odd := slice.Partition(func(n int) bool {
		return n%2 == 0
	})

	expectedEven := []int{2, 4, 6}
	expectedOdd := []int{1, 3, 5}

	if !reflect.DeepEqual(even.ToSlice(), expectedEven) {
		t.Errorf("Slice.Partition() even = %v, want %v", even.ToSlice(), expectedEven)
	}
	if !reflect.DeepEqual(odd.ToSlice(), expectedOdd) {
		t.Errorf("Slice.Partition() odd = %v, want %v", odd.ToSlice(), expectedOdd)
	}
}

func TestSliceSort(t *testing.T) {
	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6}
	slice := NewSlice(numbers)
	slice.Sort(func(a, b int) bool {
		return a < b
	})
	expected := []int{1, 1, 2, 3, 4, 5, 6, 9}
	if !reflect.DeepEqual(slice.ToSlice(), expected) {
		t.Errorf("Slice.Sort() = %v, want %v", slice.ToSlice(), expected)
	}
}

func TestSliceSortBy(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	people := []Person{
		{"Charlie", 30},
		{"Alice", 25},
		{"Bob", 30},
		{"David", 25},
	}

	slice := NewSlice(people)
	SortBy(slice, func(p Person) string {
		return p.Name
	})

	expected := []Person{
		{"Alice", 25},
		{"Bob", 30},
		{"Charlie", 30},
		{"David", 25},
	}

	if !reflect.DeepEqual(slice.ToSlice(), expected) {
		t.Errorf("Slice.SortBy() = %v, want %v", slice.ToSlice(), expected)
	}
}

func TestSliceFirst(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	first, exists := slice.First()
	if !exists || first != 1 {
		t.Errorf("Slice.First() = %v, %v, want 1, true", first, exists)
	}

	emptySlice := NewSlice([]int{})
	_, exists = emptySlice.First()
	if exists {
		t.Errorf("Slice.First() should return false for empty slice")
	}
}

func TestSliceLast(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	last, exists := slice.Last()
	if !exists || last != 5 {
		t.Errorf("Slice.Last() = %v, %v, want 5, true", last, exists)
	}

	emptySlice := NewSlice([]int{})
	_, exists = emptySlice.Last()
	if exists {
		t.Errorf("Slice.Last() should return false for empty slice")
	}
}

func TestSliceCount(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6}
	slice := NewSlice(numbers)
	count := slice.Count(func(n int) bool {
		return n%2 == 0
	})
	if count != 3 {
		t.Errorf("Slice.Count() = %v, want 3", count)
	}
}

func TestSliceAny(t *testing.T) {
	numbers := []int{1, 3, 5, 7, 9}
	slice := NewSlice(numbers)
	if !slice.Any(func(n int) bool {
		return n > 5
	}) {
		t.Errorf("Slice.Any() should return true")
	}

	if slice.Any(func(n int) bool {
		return n > 10
	}) {
		t.Errorf("Slice.Any() should return false")
	}
}

func TestSliceAll(t *testing.T) {
	numbers := []int{2, 4, 6, 8, 10}
	slice := NewSlice(numbers)
	if !slice.All(func(n int) bool {
		return n%2 == 0
	}) {
		t.Errorf("Slice.All() should return true")
	}

	if slice.All(func(n int) bool {
		return n > 5
	}) {
		t.Errorf("Slice.All() should return false")
	}
}

func TestSliceConcat(t *testing.T) {
	slice1 := NewSlice([]int{1, 2, 3})
	slice2 := NewSlice([]int{4, 5, 6})

	concatenated := slice1.Concat(slice2)
	expected := []int{1, 2, 3, 4, 5, 6}
	if !reflect.DeepEqual(concatenated.ToSlice(), expected) {
		t.Errorf("Slice.Concat() = %v, want %v", concatenated.ToSlice(), expected)
	}
}

func TestSliceIntersect(t *testing.T) {
	slice1 := NewSlice([]int{1, 2, 3, 4, 5})
	slice2 := NewSlice([]int{3, 4, 5, 6, 7})

	intersection := Intersect(slice1, slice2)
	expected := []int{3, 4, 5}
	if !reflect.DeepEqual(intersection.ToSlice(), expected) {
		t.Errorf("Slice.Intersect() = %v, want %v", intersection.ToSlice(), expected)
	}
}

func TestSliceDifference(t *testing.T) {
	slice1 := NewSlice([]int{1, 2, 3, 4, 5})
	slice2 := NewSlice([]int{3, 4, 5, 6, 7})

	difference := Difference(slice1, slice2)
	expected := []int{1, 2}
	if !reflect.DeepEqual(difference.ToSlice(), expected) {
		t.Errorf("Slice.Difference() = %v, want %v", difference.ToSlice(), expected)
	}
}

func TestSliceUnion(t *testing.T) {
	slice1 := NewSlice([]int{1, 2, 3, 4, 5})
	slice2 := NewSlice([]int{3, 4, 5, 6, 7})

	union := Union(slice1, slice2)
	expected := []int{1, 2, 3, 4, 5, 6, 7}
	if !reflect.DeepEqual(union.ToSlice(), expected) {
		t.Errorf("Slice.Union() = %v, want %v", union.ToSlice(), expected)
	}
}

func TestSliceForEach(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	sum := 0
	slice.ForEach(func(n int) {
		sum += n
	})
	expected := 15
	if sum != expected {
		t.Errorf("Slice.ForEach() sum = %v, want %v", sum, expected)
	}
}

func TestSliceForEachWithIndex(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	slice := NewSlice(numbers)
	sum := 0
	slice.ForEachWithIndex(func(i, n int) {
		sum += i + n
	})
	expected := 25 // (0+1) + (1+2) + (2+3) + (3+4) + (4+5) = 1+3+5+7+9 = 25
	if sum != expected {
		t.Errorf("Slice.ForEachWithIndex() sum = %v, want %v", sum, expected)
	}
}

// 测试便捷函数（保持向后兼容）
func TestFilter(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6}
	even := SliceFilter(numbers, func(n int) bool {
		return n%2 == 0
	})
	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(even, expected) {
		t.Errorf("Filter() = %v, want %v", even, expected)
	}
}

func TestMap(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	squares := SliceMap(numbers, func(n int) int {
		return n * n
	})
	expected := []int{1, 4, 9, 16, 25}
	if !reflect.DeepEqual(squares, expected) {
		t.Errorf("Map() = %v, want %v", squares, expected)
	}
}

func TestReduce(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	sum := SliceReduce(numbers, 0, func(acc, n int) int {
		return acc + n
	})
	expected := 15
	if sum != expected {
		t.Errorf("Reduce() = %v, want %v", sum, expected)
	}
}

func TestFind(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	found, exists := SliceFind(numbers, func(n int) bool {
		return n > 3
	})
	if !exists || found != 4 {
		t.Errorf("Find() = %v, %v, want 4, true", found, exists)
	}

	_, exists = SliceFind(numbers, func(n int) bool {
		return n > 10
	})
	if exists {
		t.Errorf("Find() should not find element > 10")
	}
}

func TestFindIndex(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	index := SliceFindIndex(numbers, func(n int) bool {
		return n == 3
	})
	if index != 2 {
		t.Errorf("FindIndex() = %v, want 2", index)
	}

	index = SliceFindIndex(numbers, func(n int) bool {
		return n == 10
	})
	if index != -1 {
		t.Errorf("FindIndex() = %v, want -1", index)
	}
}

func TestContains(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	if !SliceContains(numbers, 3) {
		t.Errorf("Contains() should find 3")
	}
	if SliceContains(numbers, 6) {
		t.Errorf("Contains() should not find 6")
	}
}

func TestIndexOf(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	index := SliceIndexOf(numbers, 3)
	if index != 2 {
		t.Errorf("IndexOf() = %v, want 2", index)
	}

	index = SliceIndexOf(numbers, 6)
	if index != -1 {
		t.Errorf("IndexOf() = %v, want -1", index)
	}
}

func TestUnique(t *testing.T) {
	numbers := []int{1, 2, 2, 3, 3, 3, 4}
	unique := SliceUnique(numbers)
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(unique, expected) {
		t.Errorf("Unique() = %v, want %v", unique, expected)
	}
}

func TestReverse(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	reversed := SliceReverse(numbers)
	expected := []int{5, 4, 3, 2, 1}
	if !reflect.DeepEqual(reversed, expected) {
		t.Errorf("Reverse() = %v, want %v", reversed, expected)
	}
}

func TestChunk(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	chunks := SliceChunk(numbers, 3)
	expected := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	if !reflect.DeepEqual(chunks, expected) {
		t.Errorf("Chunk() = %v, want %v", chunks, expected)
	}

	// Test with remainder
	numbers = []int{1, 2, 3, 4, 5}
	chunks = SliceChunk(numbers, 2)
	expected = [][]int{{1, 2}, {3, 4}, {5}}
	if !reflect.DeepEqual(chunks, expected) {
		t.Errorf("Chunk() = %v, want %v", chunks, expected)
	}
}

func TestFlatten(t *testing.T) {
	matrix := [][]int{{1, 2}, {3, 4}, {5, 6}}
	flattened := SliceFlatten(matrix)
	expected := []int{1, 2, 3, 4, 5, 6}
	if !reflect.DeepEqual(flattened, expected) {
		t.Errorf("Flatten() = %v, want %v", flattened, expected)
	}
}

func TestTake(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	taken := SliceTake(numbers, 3)
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(taken, expected) {
		t.Errorf("Take() = %v, want %v", taken, expected)
	}

	// Test taking more than available
	taken = SliceTake(numbers, 10)
	if !reflect.DeepEqual(taken, numbers) {
		t.Errorf("Take() = %v, want %v", taken, numbers)
	}
}

func TestDrop(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	dropped := SliceDrop(numbers, 2)
	expected := []int{3, 4, 5}
	if !reflect.DeepEqual(dropped, expected) {
		t.Errorf("Drop() = %v, want %v", dropped, expected)
	}

	// Test dropping more than available
	dropped = SliceDrop(numbers, 10)
	expected = []int{}
	if !reflect.DeepEqual(dropped, expected) {
		t.Errorf("Drop() = %v, want %v", dropped, expected)
	}
}

func TestTakeWhile(t *testing.T) {
	numbers := []int{2, 4, 6, 7, 8, 9}
	taken := SliceTakeWhile(numbers, func(n int) bool {
		return n%2 == 0
	})
	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(taken, expected) {
		t.Errorf("TakeWhile() = %v, want %v", taken, expected)
	}
}

func TestDropWhile(t *testing.T) {
	numbers := []int{2, 4, 6, 7, 8, 9}
	dropped := SliceDropWhile(numbers, func(n int) bool {
		return n%2 == 0
	})
	expected := []int{7, 8, 9}
	if !reflect.DeepEqual(dropped, expected) {
		t.Errorf("DropWhile() = %v, want %v", dropped, expected)
	}
}

func TestGroupBy(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	people := []Person{
		{"Alice", 25},
		{"Bob", 30},
		{"Charlie", 25},
		{"David", 30},
	}

	groups := SliceGroupBy(people, func(p Person) int {
		return p.Age
	})

	expected25 := []Person{{"Alice", 25}, {"Charlie", 25}}
	expected30 := []Person{{"Bob", 30}, {"David", 30}}

	if !reflect.DeepEqual(groups[25], expected25) {
		t.Errorf("GroupBy() age 25 = %v, want %v", groups[25], expected25)
	}
	if !reflect.DeepEqual(groups[30], expected30) {
		t.Errorf("GroupBy() age 30 = %v, want %v", groups[30], expected30)
	}
}

func TestPartition(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6}
	even, odd := SlicePartition(numbers, func(n int) bool {
		return n%2 == 0
	})

	expectedEven := []int{2, 4, 6}
	expectedOdd := []int{1, 3, 5}

	if !reflect.DeepEqual(even, expectedEven) {
		t.Errorf("Partition() even = %v, want %v", even, expectedEven)
	}
	if !reflect.DeepEqual(odd, expectedOdd) {
		t.Errorf("Partition() odd = %v, want %v", odd, expectedOdd)
	}
}

func TestSort(t *testing.T) {
	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6}
	SliceSort(numbers, func(a, b int) bool {
		return a < b
	})
	expected := []int{1, 1, 2, 3, 4, 5, 6, 9}
	if !reflect.DeepEqual(numbers, expected) {
		t.Errorf("Sort() = %v, want %v", numbers, expected)
	}
}

func TestSortBy(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	people := []Person{
		{"Charlie", 30},
		{"Alice", 25},
		{"Bob", 30},
		{"David", 25},
	}

	SliceSortBy(people, func(p Person) string {
		return p.Name
	})

	expected := []Person{
		{"Alice", 25},
		{"Bob", 30},
		{"Charlie", 30},
		{"David", 25},
	}

	if !reflect.DeepEqual(people, expected) {
		t.Errorf("SortBy() = %v, want %v", people, expected)
	}
}

func TestIsEmpty(t *testing.T) {
	if !SliceIsEmpty([]int{}) {
		t.Errorf("IsEmpty() should return true for empty slice")
	}
	if SliceIsEmpty([]int{1, 2, 3}) {
		t.Errorf("IsEmpty() should return false for non-empty slice")
	}
}

func TestFirst(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	first, exists := SliceFirst(numbers)
	if !exists || first != 1 {
		t.Errorf("First() = %v, %v, want 1, true", first, exists)
	}

	_, exists = SliceFirst([]int{})
	if exists {
		t.Errorf("First() should return false for empty slice")
	}
}

func TestLast(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	last, exists := SliceLast(numbers)
	if !exists || last != 5 {
		t.Errorf("Last() = %v, %v, want 5, true", last, exists)
	}

	_, exists = SliceLast([]int{})
	if exists {
		t.Errorf("Last() should return false for empty slice")
	}
}

func TestCount(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6}
	count := SliceCount(numbers, func(n int) bool {
		return n%2 == 0
	})
	if count != 3 {
		t.Errorf("Count() = %v, want 3", count)
	}
}

func TestAny(t *testing.T) {
	numbers := []int{1, 3, 5, 7, 9}
	if !SliceAny(numbers, func(n int) bool {
		return n > 5
	}) {
		t.Errorf("Any() should return true")
	}

	if SliceAny(numbers, func(n int) bool {
		return n > 10
	}) {
		t.Errorf("Any() should return false")
	}
}

func TestAll(t *testing.T) {
	numbers := []int{2, 4, 6, 8, 10}
	if !SliceAll(numbers, func(n int) bool {
		return n%2 == 0
	}) {
		t.Errorf("All() should return true")
	}

	if SliceAll(numbers, func(n int) bool {
		return n > 5
	}) {
		t.Errorf("All() should return false")
	}
}

func TestConcat(t *testing.T) {
	slice1 := []int{1, 2, 3}
	slice2 := []int{4, 5, 6}
	slice3 := []int{7, 8, 9}

	concatenated := SliceConcat(slice1, slice2, slice3)
	expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	if !reflect.DeepEqual(concatenated, expected) {
		t.Errorf("Concat() = %v, want %v", concatenated, expected)
	}
}

func TestIntersect(t *testing.T) {
	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := []int{3, 4, 5, 6, 7}

	intersection := SliceIntersect(slice1, slice2)
	expected := []int{3, 4, 5}
	if !reflect.DeepEqual(intersection, expected) {
		t.Errorf("Intersect() = %v, want %v", intersection, expected)
	}
}

func TestDifference(t *testing.T) {
	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := []int{3, 4, 5, 6, 7}

	difference := SliceDifference(slice1, slice2)
	expected := []int{1, 2}
	if !reflect.DeepEqual(difference, expected) {
		t.Errorf("Difference() = %v, want %v", difference, expected)
	}
}

func TestUnion(t *testing.T) {
	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := []int{3, 4, 5, 6, 7}

	union := SliceUnion(slice1, slice2)
	expected := []int{1, 2, 3, 4, 5, 6, 7}
	if !reflect.DeepEqual(union, expected) {
		t.Errorf("Union() = %v, want %v", union, expected)
	}
}

// 边界情况测试
func TestEmptySliceOperations(t *testing.T) {
	emptySlice := NewSlice([]int{})

	// 测试空切片的基础操作
	if !emptySlice.IsEmpty() {
		t.Errorf("Empty slice should be empty")
	}

	if emptySlice.Len() != 0 {
		t.Errorf("Empty slice length should be 0")
	}

	// 测试空切片的过滤操作
	filtered := emptySlice.Filter(func(n int) bool { return n > 0 })
	if !filtered.IsEmpty() {
		t.Errorf("Filtering empty slice should return empty slice")
	}

	// 测试空切片的映射操作
	mapped := emptySlice.Map(func(n int) int { return n * 2 })
	if !mapped.IsEmpty() {
		t.Errorf("Mapping empty slice should return empty slice")
	}

	// 测试空切片的归约操作
	sum := emptySlice.Reduce(0, func(acc, n int) int { return acc + n })
	if sum != 0 {
		t.Errorf("Reducing empty slice should return initial value")
	}

	// 测试空切片的查找操作
	_, found := emptySlice.Find(func(n int) bool { return n > 0 })
	if found {
		t.Errorf("Finding in empty slice should return false")
	}

	// 测试空切片的First和Last操作
	_, exists := emptySlice.First()
	if exists {
		t.Errorf("First() on empty slice should return false")
	}

	_, exists = emptySlice.Last()
	if exists {
		t.Errorf("Last() on empty slice should return false")
	}
}

func TestZeroValueOperations(t *testing.T) {
	// 测试零值切片
	var zeroSlice []int
	slice := NewSlice(zeroSlice)

	if !slice.IsEmpty() {
		t.Errorf("Zero value slice should be empty")
	}

	// 测试负数参数
	numbers := []int{1, 2, 3, 4, 5}
	slice = NewSlice(numbers)

	// 测试负数Take
	taken := slice.Take(-1)
	if !taken.IsEmpty() {
		t.Errorf("Take(-1) should return empty slice")
	}

	// 测试负数Drop
	dropped := slice.Drop(-1)
	if !reflect.DeepEqual(dropped.ToSlice(), numbers) {
		t.Errorf("Drop(-1) should return original slice")
	}

	// 测试零值Chunk
	chunks := slice.Chunk(0)
	if chunks != nil {
		t.Errorf("Chunk(0) should return nil")
	}
}
