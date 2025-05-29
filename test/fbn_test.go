package main

import (
	"context"
	"testing"
	"time"
)

func do(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	//do anything

	return nil
}

// 优化版本 1: 迭代 + 分段检查
func fibIterWithChunkCheck(ctx context.Context, n int) (int, error) {
	const checkInterval = 100 // 每100次迭代检查一次

	a, b := 0, 1
	for i := 0; i < n; i++ {
		if i%checkInterval == 0 {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			default:
			}
		}
		a, b = b, a+b
	}
	return a, nil
}

// 优化版本 2: 递归 + 深度检查
func fibRecurWithDepthCheck(ctx context.Context, n int) (int, error) {
	const checkDepth = 5 // 每5层递归检查一次

	var helper func(int, int, int, int) (int, error)
	helper = func(n, a, b, depth int) (int, error) {
		if n == 0 {
			return a, nil
		}

		if depth >= checkDepth {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			default:
				depth = 0
			}
		}

		res, err := helper(n-1, b, a+b, depth+1)
		return res, err
	}

	return helper(n, 0, 1, 0)
}

// 优化版本 3: 定时检查（混合模式）
func fibHybridCheck(ctx context.Context, n int) (int, error) {
	const checkTime = 10 * time.Millisecond

	ticker := time.NewTicker(checkTime)
	defer ticker.Stop()

	a, b := 0, 1
	for i := 0; i < n; i++ {
		select {
		case <-ticker.C:
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			default:
			}
		default:
		}
		a, b = b, a+b
	}
	return a, nil
}

// 基准测试
func BenchmarkFibIterChunkCheck20(b *testing.B)  { benchFib(b, 20, fibIterWithChunkCheck) }
func BenchmarkFibRecurDepthCheck20(b *testing.B) { benchFib(b, 20, fibRecurWithDepthCheck) }
func BenchmarkFibHybridCheck20(b *testing.B)     { benchFib(b, 20, fibHybridCheck) }

func benchFib(b *testing.B, n int, fn func(context.Context, int) (int, error)) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn(ctx, n)
	}
}
