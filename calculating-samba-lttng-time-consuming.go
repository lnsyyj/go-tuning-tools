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
	call_name string
	call_number int
	call_time_sum float64
}

var final_relust []result

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
		//fmt.Println(strEnterContent)
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

				//fmt.Println(len(lttng_exit_line_number))
				for i := 0; i < len(lttng_exit_line_number); i++ {
					strExitContent := lttng_file_content_arr[lttng_exit_line_number[i]]
					//fmt.Println(strExitContent)
					//匹配VFS名称
					matched, err := regexp.MatchString(pattern_exit, strExitContent)
					if err != nil {
						panic(err)
					}
					if matched {
						compileRegex := regexp.MustCompile(pattern_exit)
						matchArr := compileRegex.FindStringSubmatch(strExitContent)
						matchExitName = matchArr[len(matchArr) - 1]

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
								//fmt.Println(matchExitTime)

								//if len(final_relust) == 0 {
								//	add_result := result {
								//		matchExitName,
								//		1,
								//		float64(matchExitTime.Sub(matchEnterTime)),
								//	}
								//	final_relust = append(final_relust, add_result)
								//} else {
								//for _, v := range final_relust {
								//	if v.call_name == matchExitName {
								//		v.call_number += 1
								//		v.call_time_sum += float64(matchExitTime.Sub(matchEnterTime))
								//	} else {
								//		v.call_name = matchExitName
								//		v.call_number += 1
								//		v.call_time_sum += float64(matchExitTime.Sub(matchEnterTime))
								//	}
								//}
								//}

								_, ok:= result_map[matchExitName]
								if ok {
									temp := result_map[matchExitName]
									temp.call_name = matchExitName
									temp.call_number += 1
									temp.call_time_sum += float64(matchExitTime.Sub(matchEnterTime))
									result_map[matchExitName] = temp
								} else {
									temp := result {
										matchExitName,
										0,
										0,
									}
									result_map[matchExitName] = temp
								}


								reslut_time := fmt.Sprintf("%s: %f", matchExitName, float64(matchExitTime.Sub(matchEnterTime)))
								fmt.Println(reslut_time)
								//fmt.Println(matchExitName + float64(matchExitTime.Sub(matchEnterTime)))
								//fmt.Println("matchExitTime - matchEnterTime")
								//fmt.Println(float64(matchExitTime.Sub(matchEnterTime)))
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

func lttng_result_print() {
	fmt.Println("==========================")
	for _, v := range final_relust {
		fmt.Println(v.call_name, v.call_number, v.call_time_sum)
	}
}

func main() {
	var (
		filePath = kingpin.Flag(
			"file.path",
			"file path to be parsed.",
		).Default("F:/TMP/linux.samba.ceph.libcephfs.lttng.log").String()
	)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	fmt.Println(*filePath)
	if *filePath != "" {
		read_lttng_log(*filePath)
		separate_enter_and_exit()
		analyze_the_result()
		lttng_result_print()
		//fmt.Println(lttng_enter_line_number)
		//fmt.Println(lttng_exit_line_number)
	} else {
		fmt.Println("file.path is empty, there is no parseable file.")
		fmt.Println("It's ugly but works.")
	}
}