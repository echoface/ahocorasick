package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cloudflare/ahocorasick"
	anknown "github.com/anknown/ahocorasick"
	cedar "github.com/eugene-fedorenko/ahocorasick"
	"github.com/r3labs/diff/v2"
)

func PanicIfErr(err error, f string, v ...interface{}) {
	if err == nil {
		return
	}
	fmt.Printf(f, v...)
	panic(err)
}

func PanicIf(ok bool, f string, v ...interface{}) {
	if ok == false {
		return
	}
	panic(fmt.Errorf(f, v...))
}

func DistinctListStr(ori []string) []string {
	result := []string{}
	resMap := map[string]bool{}
	for _, v := range ori {
		if _, ok := resMap[v]; ok {
			continue
		}
		resMap[v] = true
		result = append(result, v)
	}
	return result
}

func main() {

	content := "休闲益智不烧脑,没wifi我也能玩一整天!,良心游戏，真送红包,红包,提现,现金,免费,红包提现,赚钱,挂机,网赚,亿万人生,刷宝,彩票,体彩,羊毛党,金币,短视频,疯狂猜成语,休闲游戏,爱上消消消,每天,明星,王二狗的摊位"

	// bad case:
	f, e := os.Open("kws.txt")
	PanicIfErr(e, "open file error")
	sc := bufio.NewScanner(f)
	keys := map[string]int{}
	for sc.Scan() {
		keys[strings.TrimSpace(sc.Text())] = len(keys)
	}
	PanicIfErr(sc.Err(), "fail read test case")
	defer f.Close()

	fmt.Println("test query content:", content)

	var cedarRound0Result []string
	var cedarRoundXResult []string

	for i := 0; i < 10; i++ {
		cedaraho := cedar.NewMatcher()

		var keyList [][]byte
		var keyRunes [][]rune
		for k, v := range keys {
			cedaraho.Insert([]byte(k), v)
			keyList = append(keyList, []byte(k))
			keyRunes = append(keyRunes, []rune(k))
		}
		cedaraho.Compile()

		var refaho *ahocorasick.Matcher
		refaho = ahocorasick.NewMatcher(keyList)
		anknownaho := &anknown.Machine{}
		err := anknownaho.Build(keyRunes)
		PanicIfErr(err, "anknownaho build err")

		fmt.Println("=========================>>>> start round n:", i)
		data := []byte(content)
		resp := cedaraho.Match(data)
		for resp.HasNext() {
			items := resp.NextMatchItem(data)
			for _, itr := range items {
				key := cedaraho.Key(data, itr)
				if i == 0 {
					cedarRound0Result = append(cedarRound0Result, string(key))
				} else {
					cedarRoundXResult = append(cedarRoundXResult, string(key))
				}
			}
		}
		cloudflareResult := []string{}
		for _, idx := range refaho.Match(data) {
			cloudflareResult = append(cloudflareResult, string(keyList[idx]))
		}
		anknownResult := []string{}
		for _, item := range anknownaho.MultiPatternSearch([]rune(content), false) {
			anknownResult = append(anknownResult, string(item.Word))
		}
		anknownResult = DistinctListStr(anknownResult)
		cloudflareResult = DistinctListStr(cloudflareResult)
		cedarRound0Result = DistinctListStr(cedarRound0Result)
		cedarRoundXResult = DistinctListStr(cedarRoundXResult)

		if i > 0 && len(cedarRound0Result) != len(cedarRoundXResult) {
			fmt.Println("anknownResult:   ", anknownResult)
			fmt.Println("cloudflareResult:", cloudflareResult)
			fmt.Println("cedaraho_round0: ", cedarRound0Result)
			fmt.Println("cedaraho_roundx: ", cedarRoundXResult)

			r, err := diff.Diff(cedarRound0Result, cedarRoundXResult)
			PanicIfErr(err, "diff run fail:%v %v",cedarRound0Result, cedarRoundXResult)
			PanicIf(true, "diff:%s", JSONPretty(r))
		}
		cedarRoundXResult = cedarRoundXResult[:0]
	}
}

func JSONPretty(v interface{}) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}
