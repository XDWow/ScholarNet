package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func longest(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	res := strs[0]
	for i := 1; i < len(strs); i++ {
		j := 0
		for ; j < len(res); j++ {
			if j > len(strs[i])-1 || res[j] != strs[i][j] {
				break
			}
		}
		res = res[:j]
		if len(res) == 0 {
			return ""
		}
	}
	return res
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		var strs []string
		if err := json.Unmarshal([]byte(line), &strs); err != nil {
			break
		}
		fmt.Println(longest(strs))
	}
}
