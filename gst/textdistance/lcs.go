package textdistance

const maxLen = 999

//Lcs ...
func Lcs(s1, s2 string) float64 {

	s1 = clearStr(s1)
	s2 = clearStr(s2)

	m := len(s1)
	n := len(s2)

	dp := [maxLen][maxLen]int{}

	for i := 0; i <= m; i++ {
		for j := 0; j <= n; j++ {
			if i == 0 || j == 0 {
				dp[i][j] = 0
			} else if string(s1[i-1]) == string(s2[j-1]) {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	nem := float64(dp[m][n])
	den := float64(m+n) / 2
	score := nem / den
	score = float64(int(score*100)) / 100
	return score
}

