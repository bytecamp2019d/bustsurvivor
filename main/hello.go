package main

import (
	"fmt"
	"math"
	"time"
)

type SQ struct {
	//矩阵结构
	M, N int //m是列数,n是行数
	Data [][]float64
}

type calcPkg struct {
	err   bool
	dur   time.Duration
	index int
}

var dyInit [][]float64          //初始时所有服务器的性能分设置为60
var scoreInit = 60.0            //初始性能分
var numServer, countTime = 4, 5 //

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
func GetRes(datay [][]float64, N int, X float64) float64 {
	Stmp := GetMatrixX()
	Sdatay := SQ{1, N + 1, datay}
	res := Mul(Stmp, Sdatay)
	sum := 0.0
	for i := 0; i <= N; i++ {
		sum += math.Pow(X, float64(i)) * res[i][0]
	}
	return sum
}
func GetMatrixX() SQ {
	var datax []float64
	for i := 0; i < countTime; i++ { //这里似乎有问题，x的取值不好说
		datax = append(datax, float64(i)*0.2+1.0)
	}
	tmp := GetX(datax, countTime-1)
	Stmp := SQ{countTime, countTime, tmp}
	return Stmp
}

func GetScore(ans calcPkg) float64 {
	erTime := 0.0
	if ans.err == true {
		erTime = 100.0
	}
	return ((ans.dur).Seconds()*1000.0 + erTime)
}
func GetServer(now calcPkg) float64 {
	var tmpY [][]float64
	for j := 1; j < countTime; j++ {
		dyInit[now.index][j-1] = dyInit[now.index][j]
		tmpY = append(tmpY, []float64{dyInit[now.index][j-1], 0})
	}
	dyInit[now.index][countTime-1] = GetScore(now)
	tmpY = append(tmpY, []float64{dyInit[now.index][countTime-1]})

	return GetRes(tmpY, countTime-1, 2)
}
func Init() {
	for i := 0; i < numServer; i++ {
		for j := 0; j < countTime; i++ {
			dyInit[i][j] = scoreInit
		}
	}
}

func getTestY() [][]float64 {
	var yy [][]float64
	for i := 1; i <= countTime; i++ {
		yy = append(yy, []float64{math.Pow(float64(i), 2) + 1.0, 0})
	}
	return yy
}
func GetDataY(dy []calcPkg) [][]float64 {
	var res [][]float64
	for i := 0; i < len(dy); i++ {
		var tt float64 = dy[i].dur.Seconds() * 1000
		er := 0.0
		if dy[i].err {
			er = 100
		}
		value := (tt + er) / 100
		res = append(res, []float64{value, 0.0})
	}
	return res
}
func main() {
	//datax := []float64{1, 2, 3, 4, 5}
	//datay := [][]float64{{2, 0}, {5, 0}, {10, 0}, {17, 0}, {26, 0}}
	//tmp := getX(datax, 4)
	//Stmp := SQ{5, 5, tmp}
	////fmt.Println(Stmp.Data)
	//Sdatay := SQ{1, 5, datay}
	//res := Mul(Stmp, Sdatay)
	//fmt.Println("res  =========", res)
	//sum := 0.0
	//for i := 0; i < 5; i++ {
	//	sum += math.Pow(6.0, float64(i)) * res[i][0]
	//}
	//fmt.Println(sum)
	//yy := getTestY()
	//fmt.Println(GetRes(yy,4,7))

}
