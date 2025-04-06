package basic_go

func findAnagrams(s string, p string) []int {
	sLen, PLen := len(s), len(p)
	res := make([]int, 0, sLen)
	if sLen < PLen {
		return nil
	}
	sCnt := [26]int{}
	pCnt := [26]int{}
	for i := 0; i < PLen; i++ {
		sCnt[s[i]-'a']++
		pCnt[p[i]-'a']++
	}
	if sCnt == pCnt {
		res = append(res, 0)
	}
	for i := 1; i <= sLen-PLen; i++ {
		// 滑动窗口左移一位，对应sCnt变化
		sCnt[s[i-1]-'a']--
		sCnt[s[i+PLen-1]-'a']++
		// 是否为字母异位词
		if sCnt == pCnt {
			res = append(res, i)
		}
	}
	return res
}
