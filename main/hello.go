package main

import (
	"fmt"
	"math"
)

type SQ struct {
	//矩阵结构
	M, N int //m是列数,n是行数
	Data [][]float64
}

//矩阵定义
func (this *SQ) Set(m int, n int, data []float64) {
	//m是列数,n是行数,data是矩阵数据（从左到右由上到下填充）
	this.M = m
	this.N = n

	k := 0
	if this.M*this.N == len(data) {
		for i := 0; i < this.N; i++ {
			var tmpArr []float64
			for j := 0; j < this.M; j++ {
				tmpArr = append(tmpArr, data[k])
				k++
			}
			this.Data = append(this.Data, tmpArr)
		}

	}
}

//a的列数和b的行数相等
//矩阵乘法
func Mul(a SQ, b SQ) [][]float64 {
	if a.M == b.N {
		res := [][]float64{}
		for i := 0; i < a.N; i++ {
			t := []float64{}
			for j := 0; j < b.M; j++ {
				r := float64(0)
				for k := 0; k < a.M; k++ {
					r += a.Data[i][k] * b.Data[k][j]
				}
				t = append(t, r)
			}
			res = append(res, t)
		}
		return res
	} else {

		return [][]float64{}
	}

}

//计算n阶行列式
func Det(Matrix [][]float64, N int) float64 {
	var T0, T1, T2, Cha int
	var Num float64
	var B [][]float64

	if N > 0 {
		Cha = 0
		for i := 0; i < N; i++ {
			var tmpArr []float64
			for j := 0; j < N; j++ {
				tmpArr = append(tmpArr, 0)
			}
			B = append(B, tmpArr)
		}
		Num = 0
		for T0 = 0; T0 <= N; T0++ {
			for T1 = 1; T1 <= N; T1++ {
				for T2 = 0; T2 <= N-1; T2++ {
					if T2 >= T0 {
						Cha = 1
					}
					B[T1-1][T2] = Matrix[T1][T2+Cha]
				} //T2循环
				Cha = 0
			} //T1循环
			Num = Num + Matrix[0][T0]*Det(B, N-1)*math.Pow(-1, float64(T0))
		} //T0循环
		return Num
	} else if N == 0 {
		return Matrix[0][0]
	}
	return 0
}

//求转置矩阵
func GetT(Matrix [][]float64, N int, M int) [][]float64 {
	var MatrixN [][]float64
	for j := 0; j <= M; j++ {
		var tmp []float64
		for i := 0; i <= N; i++ {
			tmp = append(tmp, Matrix[i][j])
		}
		MatrixN = append(MatrixN, tmp)
	}
	return MatrixN
}

//矩阵求逆
func Inverse(Matrix [][]float64, N int) (MatrixC [][]float64) {
	var T0, T1, T2, T3 int
	var B [][]float64

	MatrixC = [][]float64{}
	for i := 0; i <= N; i++ {
		var tmpArr []float64
		for j := 0; j <= N; j++ {
			tmpArr = append(tmpArr, 0)
		}
		MatrixC = append(MatrixC, tmpArr)
		B = append(B, tmpArr)
	}
	Chay := 0
	Chax := 0
	var add float64
	add = 1 / Det(Matrix, N)

	mTemp := [][]float64{}
	for T0 = 0; T0 <= N; T0++ {
		var tmp []float64
		for T3 = 0; T3 <= N; T3++ {
			for T1 = 0; T1 <= N-1; T1++ {
				if T1 < T0 {
					Chax = 0
				} else {
					Chax = 1
				}
				for T2 = 0; T2 <= N-1; T2++ {
					if T2 < T3 {
						Chay = 0
					} else {
						Chay = 1
					}
					B[T1][T2] = Matrix[T1+Chax][T2+Chay]
				}
			}
			tmp = append(tmp, Det(B, N-1)*add*(math.Pow(-1, float64(T0+T3))))
		}
		mTemp = append(mTemp, tmp)
	}
	MatrixC = GetT(mTemp, N, N)

	return MatrixC
}
func getX(x []float64, N int) [][]float64 {
	COUNT := 4
	var matrix [][]float64
	for i := 0; i <= N; i++ {
		var tmp []float64
		for j := 0; j <= COUNT; j++ {
			tmp = append(tmp, math.Pow(x[i], (float64(j))))
		}
		matrix = append(matrix, tmp)
	}
	fmt.Println("X ===== ", matrix)
	Smatrix := SQ{COUNT + 1, N + 1, matrix}
	matrixT := GetT(matrix, N, COUNT)
	fmt.Println("XT ===== ", matrixT)
	SmatrixT := SQ{N + 1, COUNT + 1, matrixT}
	tmp := Mul(SmatrixT, Smatrix)
	tmp = Inverse(tmp, N)
	fmt.Println("XT * X * -1 ===== ", tmp)
	Stmp := SQ{N + 1, N + 1, tmp}
	tmp = Mul(Stmp, SmatrixT)
	//fmt.Println(tmp)
	fmt.Println("tmp ===== ", tmp)
	return tmp
}

func main() {
	datax := []float64{1, 2, 3, 4, 5}
	datay := [][]float64{{2, 0}, {5, 0}, {10, 0}, {17, 0}, {26, 0}}
	tmp := getX(datax, 4)
	Stmp := SQ{5, 5, tmp}
	//fmt.Println(Stmp.Data)
	Sdatay := SQ{1, 5, datay}
	res := Mul(Stmp, Sdatay)
	fmt.Println("res  =========", res)
	sum := 0.0
	for i := 0; i < 5; i++ {
		sum += math.Pow(6.0, float64(i)) * res[i][0]
	}
	fmt.Println(sum)
}
