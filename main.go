package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	MAX_MEMBER_COUNT int16 = 250
	MAX_32BIT_INT    int32 = int32(math.MaxInt32)
)

type ScoreInfo struct {
	Score       int16
	MemberCount int16
	ClearDt     int32
}

func (s ScoreInfo) SaveScore() float64 {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, s.Score)
	binary.Write(buf, binary.BigEndian, MAX_MEMBER_COUNT-s.MemberCount)
	binary.Write(buf, binary.BigEndian, MAX_32BIT_INT-s.ClearDt)
	return math.Float64frombits(binary.BigEndian.Uint64(buf.Bytes()))
}

func (s ScoreInfo) ReadScore(redisScore float64) ScoreInfo {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, math.Float64bits(redisScore)) // 2^64-1까지 변환 가능
	s.Score = int16(binary.BigEndian.Uint16(buf.Bytes()[0:2]))
	s.MemberCount = MAX_MEMBER_COUNT - int16(binary.BigEndian.Uint16(buf.Bytes()[2:4]))
	s.ClearDt = MAX_32BIT_INT - int32(binary.BigEndian.Uint32(buf.Bytes()[4:8]))

	return s
}

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 서버 주소
		Password: "",               // Redis 인증 비밀번호 (비어있을 경우 없음)
		DB:       0,                // 사용할 Redis 데이터베이스 번호
	})

	// Redis 서버에 연결
	err := client.Ping(context.Background()).Err()
	if err != nil {
		panic(err)
	}

	scoreInfo := ScoreInfo{
		Score:       23346,
		MemberCount: 230,
		ClearDt:     int32(time.Now().Unix()),
	}

	fmt.Println("prevScoreInfo: ", scoreInfo)

	saveScore := scoreInfo.SaveScore()

	err = client.ZAdd(context.Background(), "RankKey4", &redis.Z{
		Score:  saveScore,
		Member: "a",
	}).Err()

	scoreInfo2 := ScoreInfo{
		Score:       32130,
		MemberCount: 134,
		ClearDt:     int32(time.Date(2023, time.June, 2, 0, 0, 0, 0, time.UTC).Unix()),
	}

	fmt.Println("prevScoreInfo2: ", scoreInfo2)

	saveScore = scoreInfo2.SaveScore()

	err = client.ZAdd(context.Background(), "RankKey4", &redis.Z{
		Score:  saveScore,
		Member: "b",
	}).Err()

	result, err := client.ZRevRangeWithScores(context.Background(), "RankKey4", 0, -1).Result()
	if err != nil {
		panic(err)
	}

	for _, z := range result {
		updateScoreInfo := scoreInfo.ReadScore(z.Score)
		fmt.Println("member:", z.Member, "afterScoreInfo:", updateScoreInfo)
	}
}
