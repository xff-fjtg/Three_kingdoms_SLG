package data

var GetYield func(rid int) Yield

// 其类型为接收 int 参数并返回 Yield 类型的函数。
var GetUnion func(rid int) int

var GetParentId func(rid int) int
var MapResTypeLevel func(x, y int) (bool, int8, int8)

var GetMainMembers func(uid int) []int
