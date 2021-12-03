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

/*

 */

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
		matched_enter, err_enter := regexp.MatchString("vfs_lttng:(.*)_enter", value)
		if err_enter != nil {
			panic(err_enter)
		}
		if matched_enter {
			lttng_enter_line_number = append(lttng_enter_line_number, lineNumber)
		}

		matched_exit, err_exit := regexp.MatchString("vfs_lttng:(.*)_exit", value)
		if err_exit != nil {
			panic(err_exit)
		}
		if matched_exit {
			lttng_exit_line_number = append(lttng_exit_line_number, lineNumber)
		}
	}
}

type result struct {
	call_number int
	call_time_sum time.Time
}

var result_map = make(map[string]result)

//var matchEnterName string
var matchExitName string
//var matchEnterTime time.Time
var matchExitTime time.Time

func analyze_the_result() {
	pattern_enter := "vfs_lttng:(.*)_enter"
	pattern_exit := "vfs_lttng:(.*)_exit"
	for _, enterValue := range lttng_enter_line_number {
		strEnterContent := lttng_file_content_arr[enterValue]
		fmt.Println(strEnterContent)
		// 匹配VFS名称
		matched, err := regexp.MatchString(pattern_enter, strEnterContent)
		if err != nil {
			panic(err)
		}
		if matched {
			compileRegex := regexp.MustCompile(pattern_enter)
			matchArr := compileRegex.FindStringSubmatch(strEnterContent)
			matchEnterName := matchArr[len(matchArr) - 1]
			//fmt.Println(matchEnterName)

			// 匹配enter时间
			matched, err = regexp.MatchString("\\[(.*)\\]", strEnterContent)
			if err != nil {
				panic(err)
			}
			if matched {
				compileRegex := regexp.MustCompile("\\[(.*)\\]")
				matchArr := compileRegex.FindStringSubmatch(strEnterContent)
				matchVFSEnterTime := matchArr[len(matchArr) - 1]
				matchEnterTime, _ := time.Parse("15:04:05", matchVFSEnterTime)
				//fmt.Println(matchEnterTime)

				fmt.Println(len(lttng_exit_line_number))
				for i := 0; i < len(lttng_exit_line_number); i++ {
					strExitContent := lttng_file_content_arr[lttng_exit_line_number[i]]
					fmt.Println(strExitContent)
					//匹配VFS名称
					matched, err := regexp.MatchString(pattern_exit, strExitContent)
					if err != nil {
						panic(err)
					}
					if matched {
						compileRegex := regexp.MustCompile(pattern_exit)
						matchArr := compileRegex.FindStringSubmatch(strExitContent)
						matchExitName = matchArr[len(matchArr) - 1]
						fmt.Println(matchExitName)
						// 匹配exit时间
						if matchEnterName == matchExitName {
							matched, err = regexp.MatchString("\\[(.*)\\]", strExitContent)
							if err != nil {
								panic(err)
							}
							if matched {
								compileRegex := regexp.MustCompile("\\[(.*)\\]")
								matchArr := compileRegex.FindStringSubmatch(strExitContent)
								matchVFSExitTime := matchArr[len(matchArr) - 1]
								//fmt.Println(matchVFSEnterTime)
								matchExitTime, _ = time.Parse("15:04:05", matchVFSExitTime)
								fmt.Println(matchExitTime)
								fmt.Println("matchExitTime - matchEnterTime")
								fmt.Println(matchExitTime.Sub(matchEnterTime))
								lttng_exit_line_number = append(lttng_exit_line_number[:i], lttng_exit_line_number[i+1:]...)
								break
							}
						}
					}
				}
			}
		} else {
			continue
		}
	}
}

func main() {
	var (
		filePath = kingpin.Flag(
			"file.path",
			"file path to be parsed.",
		).Default("F:/TMP/linux.samba.ceph.libcephfs.lttng.log").String()
	)
	kingpin.Parse()

	fmt.Println(*filePath)
	if *filePath != "" {
		read_lttng_log(*filePath)
		separate_enter_and_exit()
		analyze_the_result()
		//fmt.Println(lttng_enter_line_number)
		//fmt.Println(lttng_exit_line_number)
	} else {
		fmt.Println("file.path is empty, there is no parseable file.")
	}
}