package redis_db

import (
	"context"
	"fmt"
	"task/config"

	"github.com/redis/go-redis/v9"
)

func testZRangeByScore() {
	var err error
	// Initialize Redis client
	err = NewRedisClient(config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       1,
	})
	if err != nil {
		panic(err)
	}

	// Define the key and score range
	key := "test_producer_task_key"
	minScore := "0"
	maxScore := "100"

	// 先添加一些测试数据
	members := []redis.Z{
		{Score: 10, Member: "task1"},
		{Score: 20, Member: "task2"},
		{Score: 30, Member: "task3"},
		{Score: 90, Member: "task4"},
	}

	// 添加测试数据到Redis
	addCmd := RedisDb.ZAdd(context.Background(), key, members...)
	if err := addCmd.Err(); err != nil {
		fmt.Printf("Failed to add test data: %v\n", err)
		return
	}
	fmt.Printf("Successfully added %d members to the sorted set\n", addCmd.Val())

	// 设置查询范围参数
	rangeBy := &redis.ZRangeBy{
		Min:    minScore, // 最小分数
		Max:    maxScore, // 最大分数
		Offset: 0,        // 偏移量，从第几个开始
		Count:  10,       // 返回的最大数量
	}

	// Get the sorted set members within the specified score range
	rangeCmd := RedisDb.ZRangeByScore(context.Background(), key, rangeBy)
	results, err := rangeCmd.Result()
	if err != nil {
		fmt.Printf("Failed to get range by score: %v\n", err)
		return
	}

	fmt.Printf("Found %d members in score range %s to %s:\n", len(results), minScore, maxScore)
	for i, member := range results {
		fmt.Printf("%d: %s\n", i+1, member)
	}

	// 如果需要同时获取分数，可以使用ZRangeByScoreWithScores
	rangeWithScoresCmd := RedisDb.ZRangeByScoreWithScores(context.Background(), key, rangeBy)
	resultsWithScores, err := rangeWithScoresCmd.Result()
	if err != nil {
		fmt.Printf("Failed to get range by score with scores: %v\n", err)
		return
	}

	fmt.Printf("\nMembers with scores:\n")
	for i, z := range resultsWithScores {
		fmt.Printf("%d: %v (score: %f)\n", i+1, z.Member, z.Score)
	}

	// 清理测试数据
	delCmd := RedisDb.Del(context.Background(), key)
	if err := delCmd.Err(); err != nil {
		fmt.Printf("Failed to clean up test data: %v\n", err)
		return
	}
	fmt.Printf("Successfully deleted %d keys\n", delCmd.Val())
}
