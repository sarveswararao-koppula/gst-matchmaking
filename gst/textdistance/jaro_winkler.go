package textdistance

import (
	"math"
)

func maxDistance(len1, len2 int) int {
	maxL := max(len1, len2)
	maxDist := math.Floor(float64(maxL)/2) - 1
	return int(maxDist)
}

func jaroDistance(s1 string, s2 string) float64 {

	if s1 == s2 {
		return 1.0
	}

	len1 := len(s1)
	len2 := len(s2)

	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	maxDist := maxDistance(len1, len2)

	var (
		match  float64 = 0
		trans  float64 = 0
		result float64 = 0
	)

	var hashS1, hashS2 map[int]bool = make(map[int]bool), make(map[int]bool)

	for i := 0; i < len1; i++ {

		for j := max(0, i-maxDist); j < min(len2, i+maxDist+1); j++ {

			val1 := string(s1[i])
			val2 := string(s2[j])

			if val1 == val2 && !hashS2[j] {
				hashS1[i] = true
				hashS2[j] = true
				match++
				break
			}

		}
	}

	if match == 0 {
		return 0.0
	}

	point := 0

	for i := 0; i < len1; i++ {
		if hashS1[i] {
			for !hashS2[point] {
				point++
			}

			if string(s1[i]) != string(s2[point]) {
				trans++
			}

			point++
		}
	}

	trans /= 2

	result = match / float64(len1)
	result += match / float64(len2)
	result += (match - trans) / match

	result /= 3.0

	return result
}

//JaroWinkler ...
func JaroWinkler(s1 string, s2 string) float64 {
	s1 = clearStr(s1)
	s2 = clearStr(s2)

	jaroDist := jaroDistance(s1, s2)

	if jaroDist > 0.7 {

		prefix := 0
		for i := 0; i < min(len(s1), len(s2)); i++ {
			if string(s1[i]) == string(s2[i]) {
				prefix++
			} else {
				break
			}
		}

		prefix = min(prefix, 4)

		jaroDist += 0.1 * float64(prefix) * (1 - jaroDist)
	}

	jaroDist = float64(int(jaroDist*100)) / 100
	return jaroDist
}

