package calculator

import (
	"math"
	"time"
)

type CalcPkg struct {
	Err   bool
	Dur   time.Duration
	Index int
}

const (
	countTime = 8
	scoreInit = 10.0
)

type SQ struct {
	//矩阵结构
	M, N int //m是列数,n是行数
	Data [][]float64
}

var Stmp SQ

var dyInit [][]float64

func Mul(a SQ, b SQ) [][]float64 {
	if a.M == b.N {
		var res [][]float64
		for i := 0; i < a.N; i++ {
			var t []float64
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

func GetX(x []float64, N int) [][]float64 {
	COUNT := N
	var matrix [][]float64
	for i := 0; i <= N; i++ {
		var tmp []float64
		for j := 0; j <= COUNT; j++ {
			tmp = append(tmp, math.Pow(x[i], (float64(j))))
		}
		matrix = append(matrix, tmp)
	}
	//fmt.Println("X ===== ", matrix)
	Smatrix := SQ{COUNT + 1, N + 1, matrix}
	matrixT := GetT(matrix, N, COUNT)
	//fmt.Println("XT ===== ", matrixT)
	SmatrixT := SQ{N + 1, COUNT + 1, matrixT}
	tmp := Mul(SmatrixT, Smatrix)
	tmp = Inverse(tmp, N)
	//fmt.Println("XT * X * -1 ===== ", tmp)
	Stmp := SQ{N + 1, N + 1, tmp}
	tmp = Mul(Stmp, SmatrixT)
	//fmt.Println(tmp)
	//fmt.Println("tmp ===== ", tmp)
	return tmp
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

func GetMatrixX() SQ {
	var datax []float64
	for i := 0; i < countTime; i++ { //这里似乎有问题，x的取值不好说
		datax = append(datax, float64(i)+1)
	}
	tmp := GetX(datax, countTime-1)
	Stmp := SQ{countTime, countTime, tmp}
	return Stmp
}
func GetScore(pkg CalcPkg) float64 {
	erTime := 0.0
	if pkg.Err == true {
		erTime = 100.0
	}
	return pkg.Dur.Seconds()*1000.0 + erTime
}

func GetRes(datay [][]float64, N int, X float64) float64 {

	Sdatay := SQ{1, N + 1, datay}
	res := Mul(Stmp, Sdatay)
	sum := 0.0
	for i := 0; i <= N; i++ {
		sum += math.Pow(X, float64(i)) * res[i][0]
	}
	return sum
}

func GetServer(now CalcPkg) float64 {
	var tmpY [][]float64
	for j := 1; j < countTime; j++ {
		dyInit[now.Index][j-1] = dyInit[now.Index][j]
		tmpY = append(tmpY, []float64{dyInit[now.Index][j-1], 0})
	}
	dyInit[now.Index][countTime-1] = GetScore(now)
	tmpY = append(tmpY, []float64{dyInit[now.Index][countTime-1], 0})
	//fmt.Println("tmpY",tmpY)
	return GetRes(tmpY, countTime-1, 10)
}

func InitCalculator() {
	Stmp = GetMatrixX()
}
