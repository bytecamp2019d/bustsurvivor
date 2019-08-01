package blackjack

import (
	"fmt"
)

const (
	// totalNum of one pack of cards (exclude jokers)
	totalNum = 52
)

// Survival : probability of survival in blackjack bust is equal to numOfPossiblecases divided by NumOfAllCases
func Survival(cardsToPick, threshold int64) (numerator, denominator int64, err error) {
	// validate cardsToPick ~ [0, 52]
	if cardsToPick < 0 || cardsToPick > totalNum {
		return 0, 0, fmt.Errorf("total cards num is %d, but parameter cardsToPick is %d", totalNum, cardsToPick)
	}
	// init pokers' point, according to blackjack
	// A -> 1 (In bust-survival mode, there's no need to take Ace as eleven)
	// 2~9 -> score 2~9
	// 10, J, Q, K -> score 10
	var pokers [totalNum]int64
	for i := int64(0); i < totalNum; i++ {
		pokers[i] = i/4 + 1 // we got 4 cards for every numeric, ♤♡♢♧.
		if pokers[i] > 10 {
			pokers[i] = 10 // for J, Q, K
		}
	}
	// create dynamic programming 3-dimensional array and init dp[0][x][x] to 1
	var dp [2][totalNum + 1][]int64
	for i := 0; i < 2; i++ {
		for j := 0; j < totalNum+1; j++ {
			dp[i][j] = make([]int64, threshold+1)
		}
	}
	for j := int64(0); j <= threshold; j++ {
		for k := 0; k <= totalNum; k++ {
			dp[0][k][j] = 1
		}
	}
	// calculate possible cases via dynamic programming
	for i := int64(1); i < cardsToPick; i++ {
		for j := int64(0); j <= threshold; j++ {
			for k := totalNum; k >= 0; k-- {
				dp[i%2][k][j] = 0
				for l := k; l < totalNum; l++ {
					if j-pokers[l] < 0 {
						continue
					}
					dp[i%2][k][j] += dp[(i-1)%2][l+1][j-pokers[l]]
				}
			}
		}
	}
	// sum possible cases up
	possible := int64(0)
	for l := 0; l < totalNum; l++ {
		if threshold-pokers[l] >= 0 {
			possible += dp[(cardsToPick-1)%2][l+1][threshold-pokers[l]]
		}
	}
	// calculate total cases C(total_num, pick_num)
	total := combination(totalNum, cardsToPick)
	// reduction of a fraction
	gcd := GCD(possible, total)
	return possible / gcd, total / gcd, nil
}

// combination refers to C(x,y)
func combination(totalNum, subSetNum int64) (result int64) {
	result = 1
	for i := int64(0); i < subSetNum; i++ {
		result = result * (totalNum - i) / (i + 1)
	}
	return
}

// GCD : greatest common divisor
func GCD(x, y int64) int64 {
	z := x % y
	if z > 0 {
		return GCD(y, z)
	}
	return y
}
