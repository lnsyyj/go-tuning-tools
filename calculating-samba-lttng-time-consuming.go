package main

import (
	"bufio"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"os"
	"regexp"
	"time"
)

var lttng_file_content_arr []string
var lttng_enter_line_number []int
var lttng_exit_line_number []int

func read_lttng_log(filePath string) {
	fi, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		lttng_file_content_arr = append(lttng_file_content_arr, string(a))
	}
}

func separate_enter_and_exit() {
	for lineNumber, value := range lttng_file_content_arr {
		matched_enter, _ := regexp.MatchString("vfs_lttng:(.*)_enter", value)
		if matched_enter {
			lttng_enter_line_number = append(lttng_enter_line_number, lineNumber)
		}

		matched_exit, _ := regexp.MatchString("vfs_lttng:(.*)_exit", value)
		if matched_exit {
			lttng_exit_line_number = append(lttng_exit_line_number, lineNumber)
		}
	}
}

type result struct {
	call_name string
	call_number int
	call_time_sum float64
}

var result_map = make(map[string]result)

var result2_pattern_enter = "vfs_lttng:(.*)_enter"
var result2_pattern_exit = "vfs_lttng:(.*)_exit"
var result2_pattern_time = "\\[(.*)\\]"

func analyze_the_result2() {
	for _, enterValue := range lttng_enter_line_number {
		strEnterContent := lttng_file_content_arr[enterValue]
		// 匹配enter名字
		compileRegexEnterName := regexp.MustCompile(result2_pattern_enter)
		matchEnterNameArr := compileRegexEnterName.FindStringSubmatch(strEnterContent)
		matchEnterName := matchEnterNameArr[len(matchEnterNameArr) - 1]
		// 匹配enter时间
		compileRegexEnterTime := regexp.MustCompile(result2_pattern_time)
		matchEnterTimeArr := compileRegexEnterTime.FindStringSubmatch(strEnterContent)
		matchVFSEnterTime := matchEnterTimeArr[len(matchEnterTimeArr) - 1]
		matchEnterTime, _ := time.Parse("15:04:05", matchVFSEnterTime)
		// 将matchEnterName加入到map中
		_, ok:= result_map[matchEnterName]
		if !ok {
			temp := result {
				matchEnterName,
				0,
				0,
			}
			result_map[matchEnterName] = temp
		}

		for i := 0; i < len(lttng_exit_line_number); i++ {
			strExitContent := lttng_file_content_arr[lttng_exit_line_number[i]]
			// 匹配exit名字
			compileRegexExitName := regexp.MustCompile(result2_pattern_exit)
			matchExitNameArr := compileRegexExitName.FindStringSubmatch(strExitContent)
			matchExitName := matchExitNameArr[len(matchExitNameArr) - 1]
			// 匹配exit时间
			compileRegexExitTime := regexp.MustCompile(result2_pattern_time)
			matchExitTimeArr := compileRegexExitTime.FindStringSubmatch(strExitContent)
			matchVFSExitTime := matchExitTimeArr[len(matchExitTimeArr) - 1]
			matchExitTime, _ := time.Parse("15:04:05", matchVFSExitTime)
			//
			if matchEnterName == matchExitName {
				temp := result_map[matchEnterName]
				temp.call_name = matchEnterName
				temp.call_number += 1
				//yujiangDebug := fmt.Sprintf("%-40s - %-40s = %-40f", matchExitTime, matchEnterTime, float64(matchExitTime.Sub(matchEnterTime)))
				//fmt.Println(yujiangDebug)
				temp.call_time_sum += float64(matchExitTime.Sub(matchEnterTime))
				result_map[matchEnterName] = temp

				lttng_exit_line_number = append(lttng_exit_line_number[:i], lttng_exit_line_number[i+1:]...)
				matchEnterName = ""
				matchExitName = ""
				break
			}
		}
	}
}

func lttng_result_print() {
	fmt.Println("==========================")
	for _, v := range result_map {
		lttng_reslut_time :=  fmt.Sprintf("%-40s : %-40d : %-40f", v.call_name, v.call_number, v.call_time_sum)
		fmt.Println(lttng_reslut_time)
		//fmt.Println(v.call_name, v.call_number, v.call_time_sum)
	}
	fmt.Println("==========================")
	for _, v := range result_map {
		lttng_reslut_time := fmt.Sprintf("%-40s : %-40d : %-40f", v.call_name, v.call_number, v.call_time_sum/1000000000)
		fmt.Println(lttng_reslut_time)
	}
}

func main() {
	var (
		filePath = kingpin.Flag(
			"file.path",
			"file path to be parsed.",
		).Default("").String()
	)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	fmt.Println(*filePath)
	if *filePath != "" {
		read_lttng_log(*filePath)
		separate_enter_and_exit()
		analyze_the_result2()
		lttng_result_print()
	} else {
		fmt.Println("file.path is empty, there is no parseable file.")
		fmt.Println("It's ugly but works.")
	}
}